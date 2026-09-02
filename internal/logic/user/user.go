package user

import (
	"context"
	"fmt"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/user"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/escape"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	service.RegisterUser(New())
}

func New() *sUser {
	return &sUser{}
}

type sUser struct{}

func (s *sUser) List(ctx context.Context, req *api.ListReq) (res *api.ListRes, err error) {
	res = &api.ListRes{Page: req.Page, Size: req.Size}

	m := g.DB().Model("sys_users").Ctx(ctx).
		Where("type", "human") // 用户管理只管人类用户，AI 用户走独立入口

	if req.Keyword != "" {
		kw := "%" + escape.Like(req.Keyword) + "%"
		m = m.Where(`(username LIKE ? ESCAPE '\' OR real_name LIKE ? ESCAPE '\')`, kw, kw)
	}

	total, err := m.Count()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "获取用户数量失败")
	}
	res.Total = total

	var users []struct {
		Id        int
		Username  string
		RealName  string
		Email     string
		Type      string
		Status    int
		CreatedAt string
	}
	err = m.Fields("id, username, real_name, email, type, status, created_at").
		Page(req.Page, req.Size).
		Order("id ASC").
		Scan(&users)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "获取用户列表失败")
	}

	for _, u := range users {
		res.List = append(res.List, api.Item{
			Id:        u.Id,
			Username:  u.Username,
			RealName:  u.RealName,
			Email:     u.Email,
			Type:      u.Type,
			Status:    u.Status,
			CreatedAt: u.CreatedAt,
		})
	}
	return res, nil
}

func (s *sUser) Create(ctx context.Context, req *api.CreateReq) (id int, err error) {
	// 用户名查重
	if cnt, _ := g.DB().Model("sys_users").Where("username", req.Username).Count(); cnt > 0 {
		return 0, fmt.Errorf("用户名已存在")
	}
	// 密码复杂度
	if !hasLetterAndDigit(req.Password) {
		return 0, fmt.Errorf("密码须同时包含字母和数字")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return 0, fmt.Errorf("密码加密失败")
	}

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Exec(
			"INSERT INTO sys_users (username, password, real_name, email, type, status, created_at, updated_at) VALUES (?, ?, ?, ?, 'human', 1, ?, ?)",
			req.Username, string(hash), req.RealName, req.Email,
			time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"),
		)
		if err != nil {
			return err
		}
		lastId, _ := result.LastInsertId()
		id = int(lastId)

		// 绑定角色
		for _, roleId := range req.RoleIds {
			if _, err := tx.Exec("INSERT INTO sys_user_roles (user_id, role_id) VALUES (?, ?)", id, roleId); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "创建用户失败")
	}
	return id, nil
}

func (s *sUser) Update(ctx context.Context, req *api.UpdateReq) (err error) {
	// 不能禁用/删除 id=1 的超管
	if req.Id == 1 && req.Status != nil && *req.Status == 0 {
		return fmt.Errorf("不能禁用内置管理员账号")
	}

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		data := g.Map{"updated_at": time.Now().Format("2006-01-02 15:04:05")}
		if req.RealName != nil {
			data["real_name"] = *req.RealName
		}
		if req.Email != nil {
			data["email"] = *req.Email
		}
		if req.Status != nil {
			data["status"] = *req.Status
		}
		if _, err := tx.Model("sys_users").Where("id", req.Id).Data(data).Update(); err != nil {
			return err
		}

		// 角色重绑（全删全插）
		if req.RoleIds != nil {
			if _, err := tx.Exec("DELETE FROM sys_user_roles WHERE user_id = ?", req.Id); err != nil {
				return err
			}
			for _, roleId := range req.RoleIds {
				if _, err := tx.Exec("INSERT INTO sys_user_roles (user_id, role_id) VALUES (?, ?)", req.Id, roleId); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "更新用户失败")
	}
	return nil
}

func (s *sUser) ResetPassword(ctx context.Context, req *api.ResetPasswordReq) (err error) {
	if !hasLetterAndDigit(req.NewPassword) {
		return fmt.Errorf("密码须同时包含字母和数字")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
	if err != nil {
		return fmt.Errorf("密码加密失败")
	}
	_, err = g.DB().Model("sys_users").Ctx(ctx).
		Where("id", req.Id).
		Data(g.Map{"password": string(hash), "updated_at": time.Now().Format("2006-01-02 15:04:05")}).
		Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "重置密码失败")
	}
	// 踢掉该用户全部 token
	g.DB().Model("sys_tokens").Ctx(ctx).Where("user_id", req.Id).Delete()
	return nil
}

func (s *sUser) Delete(ctx context.Context, id int) (err error) {
	if id == 1 {
		return fmt.Errorf("不能删除内置管理员账号")
	}
	// 检查是否为某项目唯一 owner
	var ownerCnt int
	g.DB().Raw("SELECT COUNT(*) FROM project_members pm WHERE pm.user_id = ? AND pm.role = 'owner' AND pm.project_id NOT IN (SELECT project_id FROM project_members WHERE role = 'owner' AND user_id != ?)", id, id).Scan(&ownerCnt)
	if ownerCnt > 0 {
		return fmt.Errorf("该用户是 %d 个项目的唯一管理员，请先转移项目所有权", ownerCnt)
	}

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Exec("DELETE FROM sys_user_roles WHERE user_id = ?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM project_members WHERE user_id = ?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM sys_tokens WHERE user_id = ?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM sys_users WHERE id = ?", id); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "删除用户失败")
	}
	return nil
}

func hasLetterAndDigit(s string) bool {
	hasLetter, hasDigit := false, false
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

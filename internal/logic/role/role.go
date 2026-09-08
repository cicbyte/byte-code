package role

import (
	"context"
	"fmt"
	"time"
	"strconv"

	api "github.com/cicbyte/byte-code/api/v1/role"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/escape"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	service.RegisterRole(New())
}

func New() *sRole {
	return &sRole{}
}

type sRole struct{}

type dbRole struct {
	Id         int
	Name       string
	Explain    string
	IsDefault  int
	Status     string
	CreatedAt  string
}

func (s *sRole) List(ctx context.Context, req *api.ListReq) (res *api.ListRes, err error) {
	res = new(api.ListRes)
	err = g.Try(ctx, func(ctx context.Context) {
		m := g.DB().Model("sys_roles")
		if req.Name != "" {
			m = m.Where("name like ? ESCAPE '|'", "%"+escape.Like(req.Name)+"%")
		}
		if req.PageSize == 0 {
			req.PageSize = 10
		}
		if req.PageNum == 0 {
			req.PageNum = 1
		}
		total, err := m.Count()
		liberr.ErrIsNil(ctx, err, "获取角色数量失败")

		var roles []dbRole
		err = m.Page(req.PageNum, req.PageSize).Order("id asc").Scan(&roles)
		liberr.ErrIsNil(ctx, err, "获取角色列表失败")

		list := make([]api.RoleItem, 0, len(roles))
		for _, r := range roles {
			menuKeys, menuIds := s.getRoleMenus(ctx, r.Id)
			list = append(list, api.RoleItem{
				Id:         strconv.Itoa(r.Id),
				Name:       r.Name,
				Explain:    r.Explain,
				IsDefault:  r.IsDefault == 1,
				MenuIds:    menuIds,
				MenuKeys:   menuKeys,
				CreateDate: r.CreatedAt,
				Status:     r.Status,
			})
		}

		res.Page = req.PageNum
		res.PageSize = req.PageSize
		res.PageCount = total
		res.List = list
	})
	return
}

// getRoleMenus 取角色绑定的菜单（names 供展示，ids 供权限树回显）
func (s *sRole) getRoleMenus(ctx context.Context, roleId int) ([]string, []int) {
	type row struct {
		Id   int
		Name string
	}
	var rows []row
	err := g.DB().Model("sys_menus m").
		InnerJoin("sys_role_menus rm", "m.id = rm.menu_id").
		Where("rm.role_id", roleId).
		Fields("m.id, m.name").
		Order("m.id ASC").
		Scan(&rows)
	if err != nil || len(rows) == 0 {
		return []string{}, []int{}
	}
	keys := make([]string, 0, len(rows))
	ids := make([]int, 0, len(rows))
	for _, r := range rows {
		keys = append(keys, r.Name)
		ids = append(ids, r.Id)
	}
	return keys, ids
}

func (s *sRole) Create(ctx context.Context, req *api.CreateReq) (id int, err error) {
	// 角色名查重（唯一约束在迁移 41 已加，这里给友好提示）
	if cnt, _ := g.DB().Model("sys_roles").Where("name", req.Name).Count(); cnt > 0 {
		return 0, fmt.Errorf("角色名已存在")
	}

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Exec(
			"INSERT INTO sys_roles (name, `explain`, is_default, status, created_at, updated_at) VALUES (?, ?, 0, 'normal', ?, ?)",
			req.Name, req.Explain,
			time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"),
		)
		if err != nil {
			return err
		}
		lastId, _ := result.LastInsertId()
		id = int(lastId)
		for _, menuId := range req.MenuIds {
			if _, err := tx.Exec("INSERT INTO sys_role_menus (role_id, menu_id) VALUES (?, ?)", id, menuId); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "创建角色失败")
	}
	return id, nil
}

func (s *sRole) Update(ctx context.Context, req *api.UpdateReq) (err error) {
	// 内置角色（id=1 超管、id=2 普通用户）不可改名/禁用
	if req.Id <= 2 && req.Name != nil {
		return fmt.Errorf("内置角色不可修改名称")
	}
	// 超管角色的菜单是权限体系兜底（system_menu/system_role 直通），禁改防自锁
	if req.Id == 1 && req.MenuIds != nil {
		return fmt.Errorf("内置超管角色的菜单不可变更")
	}

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		data := g.Map{"updated_at": time.Now().Format("2006-01-02 15:04:05")}
		if req.Name != nil {
			data["name"] = *req.Name
		}
		if req.Explain != nil {
			data["`explain`"] = *req.Explain
		}
		if req.Status != nil {
			data["status"] = *req.Status
		}
		if _, err := tx.Model("sys_roles").Where("id", req.Id).Data(data).Update(); err != nil {
			return err
		}
		// 菜单重绑
		if req.MenuIds != nil {
			if _, err := tx.Exec("DELETE FROM sys_role_menus WHERE role_id = ?", req.Id); err != nil {
				return err
			}
			for _, menuId := range req.MenuIds {
				if _, err := tx.Exec("INSERT INTO sys_role_menus (role_id, menu_id) VALUES (?, ?)", req.Id, menuId); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "更新角色失败")
	}
	return nil
}

func (s *sRole) Delete(ctx context.Context, id int) (err error) {
	if id <= 2 {
		return fmt.Errorf("内置角色不可删除")
	}
	// 检查是否有用户绑定
	if cnt, _ := g.DB().Model("sys_user_roles").Where("role_id", id).Count(); cnt > 0 {
		return fmt.Errorf("该角色已绑定 %d 个用户，请先解除绑定", cnt)
	}

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Exec("DELETE FROM sys_role_menus WHERE role_id = ?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM sys_roles WHERE id = ?", id); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "删除角色失败")
	}
	return nil
}

func (s *sRole) UpdateMenus(ctx context.Context, id int, menuIds []int) (err error) {
	if id == 1 {
		return fmt.Errorf("内置超管角色的菜单不可变更")
	}
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Exec("DELETE FROM sys_role_menus WHERE role_id = ?", id); err != nil {
			return err
		}
		for _, menuId := range menuIds {
			if _, err := tx.Exec("INSERT INTO sys_role_menus (role_id, menu_id) VALUES (?, ?)", id, menuId); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "分配菜单权限失败")
	}
	return nil
}

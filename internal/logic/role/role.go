package role

import (
	"context"
	"strconv"

	api "github.com/cicbyte/byte-code/api/v1/role"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/escape"
	liberr "github.com/cicbyte/byte-code/library/liberr"
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
			m = m.Where("name like ? ESCAPE '\\'", "%"+escape.Like(req.Name)+"%")
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
			menuKeys := s.getRoleMenuKeys(ctx, r.Id)
			list = append(list, api.RoleItem{
				Id:         strconv.Itoa(r.Id),
				Name:       r.Name,
				Explain:    r.Explain,
				IsDefault:  r.IsDefault == 1,
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

func (s *sRole) getRoleMenuKeys(ctx context.Context, roleId int) []string {
	type nameRow struct {
		Name string
	}
	var rows []nameRow
	err := g.DB().Model("sys_menus m").
		InnerJoin("sys_role_menus rm", "m.id = rm.menu_id").
		Where("rm.role_id", roleId).
		Fields("m.name").
		Scan(&rows)
	if err != nil || len(rows) == 0 {
		return []string{}
	}
	keys := make([]string, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, row.Name)
	}
	return keys
}

package model

import "strings"

// PageReq 公共请求参数
type PageReq struct {
	DateRange []string `p:"dateRange"` //日期范围
	PageNum   int      `p:"pageNum"`   //当前页码
	PageSize  int      `p:"pageSize"`  //每页数
	OrderBy   string   `p:"orderBy" v:"regex:^$|^[a-zA-Z_][a-zA-Z0-9_]*( (asc|desc))?$#排序参数仅支持「字段名 [asc|desc]」格式"` //排序方式：字段名 [asc|desc]，可用字段由各接口白名单决定
}

// SafeOrderBy 将用户传入的排序参数约束到白名单内，返回可安全拼入 ORDER BY 的串。
// fields 的 key 为入参字段名（不区分大小写），value 为实际列名；任何非法输入一律回退 def，
// 防止排序参数被拼接进 SQL 造成注入。
func SafeOrderBy(orderBy string, fields map[string]string, def string) string {
	if orderBy == "" {
		return def
	}
	parts := strings.Fields(strings.ToLower(orderBy))
	if len(parts) == 0 || len(parts) > 2 {
		return def
	}
	column, ok := fields[parts[0]]
	if !ok {
		return def
	}
	direction := "asc"
	if len(parts) == 2 {
		if parts[1] != "asc" && parts[1] != "desc" {
			return def
		}
		direction = parts[1]
	}
	return column + " " + direction
}

// ListRes 列表公共返回
type ListRes struct {
	CurrentPage int         `json:"currentPage"`
	Total       interface{} `json:"total"`
}

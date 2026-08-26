package escape

import "strings"

// Like 转义 LIKE 模式中的通配符（% _ \），返回安全字面量；
// 调用方拼接 % 前后缀，且 SQL 须带 ESCAPE '\\' 子句（SQLite LIKE 默认无转义符）
func Like(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

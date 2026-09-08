package escape

import "strings"

// Like 转义 LIKE 模式中的通配符（% _ 及转义符 | 本身），返回安全字面量。
// 调用方拼接 % 前后缀，且 SQL 须带 ESCAPE '|' 子句。
// 转义符用 | 而非 \：MySQL 字符串字面量中反斜杠是转义字符（'\' 会吞引号破坏语法），
// | 在 SQLite/MySQL 双方言均无特殊含义，一处写法两边通用；
// 参数值经占位符绑定传输，模式中的字面量反斜杠无需再转义
func Like(s string) string {
	r := strings.NewReplacer(`|`, `||`, `%`, `|%`, `_`, `|_`)
	return r.Replace(s)
}

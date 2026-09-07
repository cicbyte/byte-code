package project

import (
	"database/sql"
	"errors"
)

// isNoRows GoFrame 对 struct 目标的 Scan 查不到行会返回 sql.ErrNoRows，
// 与真实查询错误区分开：查无此行应走调用方的零值兜底分支（"xx不存在"），
// 而不是被 WrapDb 误包装成数据库故障（如"查询任务失败"）。
// 参考 internal/logic/auth/auth.go 登录处的同类处理。
func isNoRows(err error) bool {
	return err != nil && errors.Is(err, sql.ErrNoRows)
}

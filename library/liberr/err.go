package liberr

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

func ErrIsNil(ctx context.Context, err error, msg ...string) {
	if !g.IsNil(err) {
		if len(msg) > 0 {
			g.Log().Error(ctx, err.Error())
			panic(msg[0])
		} else {
			panic(err.Error())
		}
	}
}

func ValueIsNil(value interface{}, msg string) {
	if g.IsNil(value) {
		panic(msg)
	}
}
// WrapDb 包装可能携带敏感细节的底层错误：错误信息含 SQL 语句/约束冲突等
// 数据库细节时只记服务端日志、对外返回通用文案；业务错误（无 SQL 特征）
// 保持"文案: 原因"透传，避免把"该文档下有子文档"这类有用信息也抹掉
func WrapDb(ctx context.Context, err error, msg string) error {
	if err == nil {
		return nil
	}
	detail := err.Error()
	upper := strings.ToUpper(detail)
	for _, kw := range []string{
		"INSERT INTO", "UPDATE ", "DELETE FROM", "SELECT ",
		"CREATE TABLE", "DROP TABLE", "ALTER TABLE",
		"CONSTRAINT FAILED", "SQL:", "SQLITE_",
	} {
		if strings.Contains(upper, kw) {
			g.Log().Errorf(ctx, "%s: %v", msg, err)
			return fmt.Errorf("%s", msg)
		}
	}
	return fmt.Errorf("%s: %v", msg, err)
}

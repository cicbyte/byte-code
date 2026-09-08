package dbutil

import (
	"context"
	"fmt"

	"github.com/cicbyte/byte-code/utility/dbinit"
	"github.com/gogf/gf/v2/frame/g"
)

// UpsertConfig 单键写入 sys_config（存在更新/不存在插入）。
// 方言分支替代 count-then-insert：消除竞态（并发首写撞主键）并让错误
// 正常冒泡——此前三处调用方的错误全部被吞，保存失败用户无感知。
func UpsertConfig(ctx context.Context, key, value string) error {
	var err error
	if dbinit.Dialect() == "mysql" {
		_, err = g.DB().Exec(ctx,
			"INSERT INTO sys_config (`key`, `value`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `value` = VALUES(`value`)",
			key, value)
	} else {
		_, err = g.DB().Exec(ctx,
			"INSERT INTO sys_config (`key`, `value`) VALUES (?, ?) ON CONFLICT(`key`) DO UPDATE SET `value` = excluded.value",
			key, value)
	}
	if err != nil {
		return fmt.Errorf("写入配置 %s 失败: %w", key, err)
	}
	return nil
}

package main

import (
	_ "github.com/cicbyte/byte-code/internal/logic"
	_ "github.com/cicbyte/byte-code/internal/packed"

	// 数据库驱动：SQLite（默认，单机零运维）与 MySQL（公网部署/数据共享）
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"github.com/cicbyte/byte-code/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

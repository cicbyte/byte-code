package main

import (
	_ "github.com/cicbyte/byte-code/internal/logic"
	_ "github.com/cicbyte/byte-code/internal/packed"

	// SQLite数据库驱动
	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"github.com/cicbyte/byte-code/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

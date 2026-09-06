package version

// Version 构建时注入（CI: -ldflags "-X github.com/cicbyte/byte-code/internal/version.Version=x.y.z"）；
// 本地 `go run` / `go build` 未注入时显示 dev
var Version = "dev"

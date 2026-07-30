package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v5"
)

// jwtSecretFile 首次启动自动生成的随机秘钥落盘位置，随数据文件一同保存
const jwtSecretFile = "resource/data/jwt.secret"

var (
	jwtSecretOnce sync.Once
	jwtSecret     []byte
)

// loadJwtSecret 加载签名秘钥，优先级：环境变量 JWT_SECRET > 配置 token.secret > 秘钥文件 > 首次随机生成并落盘。
// 秘钥不可硬编码在源码中，否则任何拿到源码的人都能伪造任意用户（含 admin）的 token。
func loadJwtSecret() []byte {
	jwtSecretOnce.Do(func() {
		if s := os.Getenv("JWT_SECRET"); s != "" {
			jwtSecret = []byte(s)
			return
		}
		if v, _ := g.Cfg().Get(context.Background(), "token.secret"); v != nil && v.String() != "" {
			jwtSecret = []byte(v.String())
			return
		}
		if b, err := os.ReadFile(jwtSecretFile); err == nil && len(b) >= 32 {
			jwtSecret = b
			return
		}
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			panic(fmt.Errorf("生成 JWT 秘钥失败: %w", err))
		}
		secret := []byte(hex.EncodeToString(b))
		// 落盘失败不阻断启动，但重启后已发 token 全部失效，用户需重新登录
		_ = os.MkdirAll(filepath.Dir(jwtSecretFile), 0o700)
		_ = os.WriteFile(jwtSecretFile, secret, 0o600)
		jwtSecret = secret
	})
	return jwtSecret
}

type Claims struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(userId int, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserId:   userId,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "byte-code",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(loadJwtSecret())
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return loadJwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

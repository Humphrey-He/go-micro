package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go-micro/pkg/config"
	"go-micro/pkg/errx"
)

const (
	CtxUserID = "user_id"
	CtxRole   = "role"
	CtxName   = "username"
)

// jwtBlacklist 未初始化（nil）时 JWTAuth 不做黑名单检查，
// 因此未接入 Redis 的服务可以继续单独使用该中间件。
var jwtBlacklist *redis.Client

// InitJWTBlacklist 注入 Redis 客户端以启用登出后的 token 拉黑。
func InitJWTBlacklist(rdb *redis.Client) {
	jwtBlacklist = rdb
}

// TokenBlacklistKey 由 token 的 SHA-256 摘要生成黑名单键，避免明文 token 落 Redis。
func TokenBlacklistKey(tokenDigest string) string {
	return "auth:revoked:" + tokenDigest
}

func tokenDigest(tokenStr string) string {
	sum := sha256.Sum256([]byte(tokenStr))
	return hex.EncodeToString(sum[:])
}

func JWTAuth() gin.HandlerFunc {
	secret := []byte(config.GetEnv("JWT_SECRET", "dev-secret"))
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"code": errx.CodeUnauthorized, "message": errx.MsgUnauthorized})
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"code": errx.CodeUnauthorized, "message": errx.MsgUnauthorized})
			return
		}
		if jwtBlacklist != nil {
			digest := tokenDigest(tokenStr)
			ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
			if n, err := jwtBlacklist.Exists(ctx, TokenBlacklistKey(digest)).Result(); err == nil && n > 0 {
				cancel()
				c.AbortWithStatusJSON(401, gin.H{"code": errx.CodeTokenInvalid, "message": errx.MsgTokenRevoked})
				return
			}
			cancel()
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"code": errx.CodeUnauthorized, "message": errx.MsgUnauthorized})
			return
		}
		if v, ok := claims["user_id"].(string); ok {
			c.Set(CtxUserID, v)
		}
		if v, ok := claims["role"].(string); ok {
			c.Set(CtxRole, v)
		}
		if v, ok := claims["username"].(string); ok {
			c.Set(CtxName, v)
		}
		c.Next()
	}
}

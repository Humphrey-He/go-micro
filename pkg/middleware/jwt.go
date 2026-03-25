package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go-micro/pkg/config"
	"go-micro/pkg/errx"
)

const (
	CtxUserID = "user_id"
	CtxRole   = "role"
	CtxName   = "username"
)

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

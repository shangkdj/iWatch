package middleware

import (
	"net/http"
	"strings"

	"watch-api/services"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 Authorization Header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "未提供认证令牌",
			})
			return
		}

		// 2. 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "令牌格式无效，请使用 Bearer 方案",
			})
			return
		}

		tokenString := parts[1]

		// 3. 验证 JWT
		claims, err := services.ValidateJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "令牌无效或已过期",
			})
			return
		}

		// 4. 将用户 ID 存入上下文（供后续接口使用）
		c.Set("userID", claims.UserID)

		// 5. 继续执行后续的 Handler
		c.Next()
	}
}

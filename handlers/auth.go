package handlers

import (
	"net/http"

	"watch-api/models"
	"watch-api/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// handlers/auth.go
// *gorm.DB 共享同一个数据库连接池，使用闭包传递 db 对象
// 为什么使用闭包，Gin的路由只接受 func(c *gin.Context) 作为处理函数，
// 而我们需要在处理函数中使用数据库连接，所以通过闭包将 db 传递进去。
func TestLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求中获取 user_id
		// 这里假设前端发送的 JSON 数据格式为 {"user_id": "some_user_id"}
		// binding:"required" 表示 user_id 是必填字段，如果缺失会返回 400 错误
		var req struct {
			UserID string `json:"user_id" binding:"required"`
		}
		//ShouldBindJSON 会自动解析请求体中的 JSON 数据，并将其绑定到 req 结构体中
		// 执行成功则返回空
		// 如果解析失败（例如 JSON 格式错误或缺少必填字段），会返回错误
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 查找或创建用户
		// 开发/测试环境下，使用 FirstOrCreate 方法，如果用户不存在则创建新用户
		// 生产环境下，应该有更严格的用户验证逻辑，例如通过数据库查询用户表，验证用户身份等
		// gin.H{"key": "value"} 等价于 map[string]interface{}{"key": "value"}
		var user models.User
		result := db.FirstOrCreate(&user, models.User{UserID: req.UserID})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "处理用户失败"})
			return
		}

		// 生成 JWT 令牌
		token, err := services.GenerateJWT(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
			return
		}

		// 返回 token 和用户信息
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":      user.ID,
				"user_id": user.UserID,
			},
		})

		// 生产环境改为结构体
		// c.JSON(http.StatusOK, models.Response{
		//     Code:    200,
		//     Message: "登录成功",
		//     Data: LoginResponse{
		//         Token: token,
		//         User: UserInfo{
		//             ID:     user.ID,
		//             UserID: user.UserID,
		//         },
		//     },
		// })
	}
}

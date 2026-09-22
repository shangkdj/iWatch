package handlers

import (
	"context"
	"net/http"
	"watch-api/models"
	"watch-api/services"

	"github.com/Timothylock/go-signin-with-apple/apple"

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

// AppleLoginRequest 苹果登录请求体
type AppleLoginRequest struct {
	//required 必填
	//omitempty 可选
	IdentityToken string `json:"identity_token" binding:"required"`
	// 以下字段苹果仅在首次授权时返回，客户端首次登录时需一并提交
	FullName string `json:"full_name,omitempty"`
	Email    string `json:"email,omitempty"`
}

// AppleLogin 处理苹果登录
// 1. 用户登录
// 2. 苹果返回identityToken（JWT）
// 3. iOS端将identityToken发送到后端
// 4. 后端验证identityToken的合法性（签名、过期时间、aud、iss等）
// 5. 后端根据苹果返回的sub（唯一标识）查找或创建用户
// 6. 后端生成自己的JWT返回给客户端

func AppleLogin(db *gorm.DB, clientID string) gin.HandlerFunc {

	return func(c *gin.Context) {
		var req AppleLoginRequest
		//shouldBindJSON 会自动解析请求体中的 JSON 数据，并将其绑定到 req 结构体中

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 1. 验证 identityToken
		//1. 解析 identityToken（JWT 格式）
		//2. 从苹果服务器获取公钥（https://appleid.apple.com/auth/keys）
		//3. 验证签名是否正确
		//4. 验证 iss 是否为 https://appleid.apple.com
		//5. 验证 aud 是否为你的 clientID（Bundle ID）
		//6. 验证 exp 是否未过期
		// 验证失败会返回401 Unauthorized
		client := apple.New()
		claims, err := client.VerifyIDToken(context.Background(), req.IdentityToken, clientID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "苹果登录验证失败: " + err.Error()})
			return
		}

		// 2. 获取苹果用户的唯一标识 sub
		appleUserID := claims.Subject // 这是苹果返回的稳定唯一标识
		if appleUserID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户标识"})
			return
		}

		// 3. 查找或创建用户（使用 FirstOrCreate 避免并发冲突）
		var user models.User
		// 注意：这里用 appleUserID 来匹配 User 表的 UserID 字段
		result := db.Where("user_id = ?", appleUserID).FirstOrCreate(&user, models.User{
			UserID: appleUserID,
			Email:  &req.Email, // 首次登录时苹果会返回
			Name:   &req.FullName,
		})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "处理用户失败: " + result.Error.Error()})
			return
		}

		// 4. 生成我们自己的 JWT
		token, err := services.GenerateJWT(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
			return
		}

		// 5. 返回 JWT 和用户信息
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":      user.ID,
				"user_id": user.UserID,
			},
		})
	}
}

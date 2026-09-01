package handlers

import (
	"net/http"

	"watch-api/models"
	"watch-api/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// handlers/auth.go
func TestLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求中获取 user_id
		var req struct {
			UserID string `json:"user_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 查找或创建用户
		var user models.User
		result := db.Where("user_id = ?", req.UserID).First(&user)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				user = models.User{UserID: req.UserID}
				db.Create(&user)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败"})
				return
			}
		}

		// 生成 JWT
		token, err := services.GenerateJWT(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":      user.ID,
				"user_id": user.UserID,
			},
		})
	}
}

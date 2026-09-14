package handlers

import (
	"net/http"

	"watch-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SaveApnsTokenRequest 保存 Token 请求
type SaveApnsTokenRequest struct {
	DeviceType string `json:"device_type" binding:"required,oneof=watch iphone"`
	Token      string `json:"token" binding:"required,min=32"`
}

// SaveApnsToken 保存或更新 APNs Token
// POST /api/v1/apns/token
func SaveApnsToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 userID
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			return
		}

		uid, ok := userID.(int64)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
			return
		}

		// 2. 解析请求
		var req SaveApnsTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 3. Upsert：如果 Token 已存在则更新
		token := models.ApnsToken{
			UserID:     uid,
			DeviceType: req.DeviceType,
			Token:      req.Token,
			IsActive:   true,
		}

		err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "token"}},
			DoUpdates: clause.AssignmentColumns([]string{"user_id", "device_type", "is_active", "updated_at"}),
		}).Create(&token).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Token 已保存",
		})
	}
}

// DeleteApnsToken 删除 APNs Token（用户关闭推送时调用）
// DELETE /api/v1/apns/token
func DeleteApnsToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			return
		}

		uid, _ := userID.(int64)

		var req struct {
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := db.Where("token = ? AND user_id = ?", req.Token, uid).
			Delete(&models.ApnsToken{}).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

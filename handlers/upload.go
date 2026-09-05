package handlers

import (
	"net/http"
	"time"

	"watch-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SampleInput 单条样本
type SampleInput struct {
	SampleUUID string    `json:"sample_uuid" binding:"required"`
	Type       string    `json:"type" binding:"required,oneof=hrv heart_rate steps"`
	Value      float64   `json:"value" binding:"required"`
	StartDate  time.Time `json:"start_date" binding:"required"`
	EndDate    time.Time `json:"end_date" binding:"required"`
}

// UploadRequest 批量上传请求
type UploadRequest struct {
	UserID  string        `json:"user_id" binding:"required"`
	Samples []SampleInput `json:"samples" binding:"required,min=1"`
}

// BatchUpload 批量上传健康数据
func BatchUpload(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UploadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 从中间件获取 userID
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

		// 组装数据
		var samples []models.HealthSample
		for _, s := range req.Samples {
			samples = append(samples, models.HealthSample{
				SampleUUID: s.SampleUUID,
				UserID:     uid,
				DataType:   s.Type,
				Value:      s.Value,
				StartDate:  s.StartDate,
				EndDate:    s.EndDate,
				Source:     "iphone",
			})
		}

		// 幂等插入
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&samples).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "插入失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"count":   len(samples),
			"message": "数据上传成功",
		})
	}
}

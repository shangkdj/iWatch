package handlers

import (
	"net/http"

	"watch-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ==================== 获取设置 ====================

// GetUserSettings 获取用户设置
// GET /api/v1/user/settings
func GetUserSettings(db *gorm.DB) gin.HandlerFunc {
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

		// 2. 查询设置
		var settings models.UserSettings
		err := db.Where("user_id = ?", uid).First(&settings).Error

		// 3. 如果不存在，返回默认设置并创建
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 创建默认设置
				settings = models.UserSettings{
					UserID:             uid,
					HrvMin:             30,
					HrvMax:             60,
					HrvAlertThreshold:  0.2,
					StepsGoal:          8000,
					WeatherPushEnabled: true,
					SleepPushEnabled:   true,
					EmotionPushEnabled: true,
					PushTime:           "07:00",
					WeatherCity:        "北京",
					UnitSystem:         "metric",
				}
				if err := db.Create(&settings).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "创建默认设置失败"})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "查询设置失败"})
				return
			}
		}

		// 4. 返回设置
		c.JSON(http.StatusOK, settings)
	}
}

// ==================== 更新设置 ====================

// UpdateSettingsRequest 更新设置请求体
type UpdateSettingsRequest struct {
	HrvMin             *float64 `json:"hrv_min"`
	HrvMax             *float64 `json:"hrv_max"`
	HrvAlertThreshold  *float64 `json:"hrv_alert_threshold" binding:"omitempty,min=0,max=1"`
	StepsGoal          *int     `json:"steps_goal" binding:"omitempty,min=1000,max=100000"`
	WeatherPushEnabled *bool    `json:"weather_push_enabled"`
	SleepPushEnabled   *bool    `json:"sleep_push_enabled"`
	EmotionPushEnabled *bool    `json:"emotion_push_enabled"`
	PushTime           *string  `json:"push_time" binding:"omitempty,len=5"`
	WeatherCity        *string  `json:"weather_city" binding:"omitempty,max=64"`
	UnitSystem         *string  `json:"unit_system" binding:"omitempty,oneof=metric imperial"`
}

// UpdateUserSettings 更新用户设置
// PUT /api/v1/user/settings
func UpdateUserSettings(db *gorm.DB) gin.HandlerFunc {
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

		// 2. 解析请求体
		var req UpdateSettingsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 3. 查找设置（不存在则创建）
		var settings models.UserSettings
		err := db.Where("user_id = ?", uid).First(&settings).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				settings = models.UserSettings{UserID: uid}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "查询设置失败"})
				return
			}
		}

		// 4. 用指针字段更新（只更新传过来的字段）
		if req.HrvMin != nil {
			settings.HrvMin = *req.HrvMin
		}
		if req.HrvMax != nil {
			settings.HrvMax = *req.HrvMax
		}
		if req.HrvAlertThreshold != nil {
			settings.HrvAlertThreshold = *req.HrvAlertThreshold
		}
		if req.StepsGoal != nil {
			settings.StepsGoal = *req.StepsGoal
		}
		if req.WeatherPushEnabled != nil {
			settings.WeatherPushEnabled = *req.WeatherPushEnabled
		}
		if req.SleepPushEnabled != nil {
			settings.SleepPushEnabled = *req.SleepPushEnabled
		}
		if req.EmotionPushEnabled != nil {
			settings.EmotionPushEnabled = *req.EmotionPushEnabled
		}
		if req.PushTime != nil {
			settings.PushTime = *req.PushTime
		}
		if req.WeatherCity != nil {
			settings.WeatherCity = *req.WeatherCity
		}
		if req.UnitSystem != nil {
			settings.UnitSystem = *req.UnitSystem
		}

		// 5. 保存（存在则更新，不存在则插入）
		if err := db.Save(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存设置失败"})
			return
		}

		// 6. 返回更新后的设置
		c.JSON(http.StatusOK, settings)
	}
}

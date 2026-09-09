// handlers/dashboard.go
package handlers

import (
	"net/http"
	"time"

	"watch-api/models"
	"watch-api/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetDashboard 首页数据接口
func GetDashboard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从中间件获取 userID
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

		// 2. 获取今日日期
		today := time.Now().Format("2006-01-02")
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

		// 3. 查询今日数据
		var todayAgg models.DailyAgg
		todayErr := db.Where("user_id = ? AND date = ?", uid, today).First(&todayAgg).Error

		// 4. 查询昨日数据（用于趋势对比）
		var yesterdayAgg models.DailyAgg
		yesterdayErr := db.Where("user_id = ? AND date = ?", uid, yesterday).First(&yesterdayAgg).Error

		// 5. 如果没有今日数据，返回空状态
		if todayErr != nil {
			c.JSON(http.StatusOK, gin.H{
				"has_data": false,
				"message":  "今日暂无数据",
				"date":     today,
			})
			return
		}

		hasYesterday := yesterdayErr == nil

		// 6. 百分比计算趋势（阈值 10%）
		thresholdPercent := 10.0

		hrvTrend := services.CalculatePercentageTrend(
			todayAgg.AvgHrv,
			yesterdayAgg.AvgHrv,
			hasYesterday,
			thresholdPercent,
		)

		heartTrend := services.CalculatePercentageTrend(
			todayAgg.AvgHeartRate,
			yesterdayAgg.AvgHeartRate,
			hasYesterday,
			thresholdPercent,
		)

		stepsTrend := services.CalculateStepTrend(
			todayAgg.TotalSteps,
			yesterdayAgg.TotalSteps,
			hasYesterday,
			thresholdPercent,
		)

		// 7. 步数目标进度（默认 8000 步）
		stepsGoal := 8000
		completion := float64(todayAgg.TotalSteps) / float64(stepsGoal) * 100
		if completion > 100 {
			completion = 100
		}

		// 8. 返回响应
		c.JSON(http.StatusOK, gin.H{
			"has_data": true,
			"date":     today,
			"hrv": gin.H{
				"value": todayAgg.AvgHrv,
				"unit":  "ms",
				"trend": hrvTrend,
			},
			"heart_rate": gin.H{
				"value": todayAgg.AvgHeartRate,
				"unit":  "bpm",
				"trend": heartTrend,
			},
			"steps": gin.H{
				"value":      todayAgg.TotalSteps,
				"goal":       stepsGoal,
				"completion": completion,
				"trend":      stepsTrend,
			},
		})
	}
}

package handlers

import (
	"net/http"
	"strconv"
	"time"

	"watch-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TrendResponse 单日趋势数据
type TrendResponse struct {
	Date       string   `json:"date"`
	AvgHrv     *float64 `json:"avg_hrv"`
	AvgHeart   *float64 `json:"avg_heart_rate"`
	TotalSteps int64    `json:"total_steps"`
}

// GetHealthTrend 获取历史趋势数据
// GET /api/v1/health/trend?days=7
func GetHealthTrend(db *gorm.DB) gin.HandlerFunc {
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

		// 2. 解析 days 参数（默认 7 天，最大 90 天）
		days := 7
		if d := c.Query("days"); d != "" {
			if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
				days = parsed
				if days > 90 {
					days = 90
				}
			}
		}

		// 3. 计算起始日期
		startDate := time.Now().AddDate(0, 0, -days+1).Format("2006-01-02")
		endDate := time.Now().Format("2006-01-02")

		// 4. 查询 daily_agg 表
		var aggs []models.DailyAgg
		err := db.Where("user_id = ? AND date >= ? AND date <= ?", uid, startDate, endDate).
			Order("date ASC").
			Find(&aggs).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		// 5. 组装响应数据
		data := make([]TrendResponse, 0, len(aggs))
		for _, agg := range aggs {
			data = append(data, TrendResponse{
				Date:       agg.Date.Format("2006-01-02"),
				AvgHrv:     agg.AvgHrv,
				AvgHeart:   agg.AvgHeartRate,
				TotalSteps: agg.TotalSteps,
			})
		}

		// 6. 返回响应
		c.JSON(http.StatusOK, gin.H{
			"total_days": len(data),
			"start_date": startDate,
			"end_date":   endDate,
			"data":       data,
		})
	}
}

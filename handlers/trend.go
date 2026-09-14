package handlers

import (
	"math"
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

// TrendSummary 时段统计摘要
type TrendSummary struct {
	// 平均值
	AvgHrv   *float64 `json:"avg_hrv"`
	AvgHeart *float64 `json:"avg_heart_rate"`
	AvgSteps float64  `json:"avg_steps"`
	// 极值
	MaxHrv   *float64 `json:"max_hrv"`
	MinHrv   *float64 `json:"min_hrv"`
	MaxHeart *float64 `json:"max_heart_rate"`
	MinHeart *float64 `json:"min_heart_rate"`
	MaxSteps int64    `json:"max_steps"`
	MinSteps int64    `json:"min_steps"`
	// 总计
	TotalSteps int64 `json:"total_steps"`
	TotalDays  int   `json:"total_days"`
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
			//atoi 把字符串转换为整数，如果转换失败会返回错误
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

// calculateSummary 计算时段统计摘要
func calculateSummary(aggs []models.DailyAgg) TrendSummary {
	summary := TrendSummary{}

	if len(aggs) == 0 {
		return summary
	}

	var hrvValues []float64
	var heartValues []float64
	var totalSteps int64
	var maxHrv, minHrv, maxHeart, minHeart *float64
	var maxSteps, minSteps int64

	for i, agg := range aggs {
		// HRV
		if agg.AvgHrv != nil {
			hrvValues = append(hrvValues, *agg.AvgHrv)
			if maxHrv == nil || *agg.AvgHrv > *maxHrv {
				maxHrv = agg.AvgHrv
			}
			if minHrv == nil || *agg.AvgHrv < *minHrv {
				minHrv = agg.AvgHrv
			}
		}

		// 心率
		if agg.AvgHeartRate != nil {
			heartValues = append(heartValues, *agg.AvgHeartRate)
			if maxHeart == nil || *agg.AvgHeartRate > *maxHeart {
				maxHeart = agg.AvgHeartRate
			}
			if minHeart == nil || *agg.AvgHeartRate < *minHeart {
				minHeart = agg.AvgHeartRate
			}
		}

		// 步数
		totalSteps += agg.TotalSteps
		if i == 0 || agg.TotalSteps > maxSteps {
			maxSteps = agg.TotalSteps
		}
		if i == 0 || agg.TotalSteps < minSteps {
			minSteps = agg.TotalSteps
		}
	}

	// 平均值
	if len(hrvValues) > 0 {
		avg := average(hrvValues)
		summary.AvgHrv = &avg
	}
	if len(heartValues) > 0 {
		avg := average(heartValues)
		summary.AvgHeart = &avg
	}
	summary.AvgSteps = math.Round(float64(totalSteps)/float64(len(aggs))*100) / 100

	// 极值
	summary.MaxHrv = maxHrv
	summary.MinHrv = minHrv
	summary.MaxHeart = maxHeart
	summary.MinHeart = minHeart
	summary.MaxSteps = maxSteps
	summary.MinSteps = minSteps

	// 总计
	summary.TotalSteps = totalSteps
	summary.TotalDays = len(aggs)

	return summary
}

// average 计算平均值
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return math.Round(sum/float64(len(values))*100) / 100
}

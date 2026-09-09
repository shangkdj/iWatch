package services

import (
	"context"
	"encoding/json"
	"time"

	"watch-api/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"gorm.io/gorm/clause"
)

// UpdateDailyAggregation 异步更新日聚合数据
func UpdateDailyAggregation(db *gorm.DB, rdb *redis.Client, samples []models.HealthSample) {
	if len(samples) == 0 {
		return
	}

	// 按用户分组
	userMap := make(map[int64][]models.HealthSample)
	for _, s := range samples {
		userMap[s.UserID] = append(userMap[s.UserID], s)
	}

	for userID, userSamples := range userMap {
		// 按日期分组
		dateMap := make(map[string][]models.HealthSample)
		for _, s := range userSamples {
			date := s.StartDate.Format("2006-01-02")
			dateMap[date] = append(dateMap[date], s)
		}

		for date, daySamples := range dateMap {
			// 计算聚合值
			agg := calculateAggregation(userID, date, daySamples)

			// 写入数据库（Upsert）
			upsertDailyAgg(db, agg)

			// 更新 Redis 缓存
			updateCache(rdb, agg)
		}
	}
}

// calculateAggregation 计算单日聚合值
func calculateAggregation(userID int64, date string, samples []models.HealthSample) models.DailyAgg {
	var agg models.DailyAgg
	agg.UserID = userID
	agg.Date, _ = time.Parse("2006-01-02", date)

	var hrvValues []float64
	var heartRateValues []float64
	var totalSteps int64

	for _, s := range samples {
		switch s.DataType {
		case "hrv":
			hrvValues = append(hrvValues, s.Value)
		case "heart_rate":
			heartRateValues = append(heartRateValues, s.Value)
		case "steps":
			totalSteps += int64(s.Value)
		}
	}

	// 计算平均值
	if len(hrvValues) > 0 {
		avg := average(hrvValues)
		agg.AvgHrv = &avg
	}
	if len(heartRateValues) > 0 {
		avg := average(heartRateValues)
		agg.AvgHeartRate = &avg
	}
	agg.TotalSteps = totalSteps

	return agg
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
	return sum / float64(len(values))
}

// upsertDailyAgg 插入或更新日聚合记录
func upsertDailyAgg(db *gorm.DB, agg models.DailyAgg) {
	// 如果所有值都是 nil/0，跳过插入
	if agg.AvgHrv == nil && agg.AvgHeartRate == nil && agg.TotalSteps == 0 {
		return
	}

	err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "date"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"avg_hrv",
			"avg_heart_rate",
			"total_steps",
			"updated_at",
		}),
	}).Create(&agg).Error

	if err != nil {
		// 记录错误但不影响主流程
		return
	}
}

// updateCache 更新 Redis 缓存
func updateCache(rdb *redis.Client, agg models.DailyAgg) {
	if rdb == nil {
		return
	}

	ctx := context.Background()
	dateStr := agg.Date.Format("2006-01-02")
	cacheKey := "comp:" + string(rune(agg.UserID)) + ":" + dateStr

	jsonData, err := json.Marshal(agg)
	if err != nil {
		return
	}

	rdb.Set(ctx, cacheKey, jsonData, 24*time.Hour)
}

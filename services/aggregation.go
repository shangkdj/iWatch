package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"watch-api/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpdateDailyAggregation 异步更新日聚合数据
func UpdateDailyAggregation(db *gorm.DB, rdb *redis.Client, samples []models.HealthSample) {
	log.Println("📊 聚合开始，样本数:", len(samples))

	if len(samples) == 0 {
		log.Println("⚠️ 样本为空，跳过聚合")
		return
	}

	// 按用户分组
	userMap := make(map[int64][]models.HealthSample)
	for _, s := range samples {
		userMap[s.UserID] = append(userMap[s.UserID], s)
	}
	log.Printf("👥 涉及用户数: %d", len(userMap))

	for userID, userSamples := range userMap {
		log.Printf("🔍 处理用户: %d, 样本数: %d", userID, len(userSamples))

		// 按日期分组
		dateMap := make(map[string][]models.HealthSample)
		for _, s := range userSamples {
			date := s.StartDate.Format("2006-01-02")
			dateMap[date] = append(dateMap[date], s)
		}
		log.Printf("📅 用户 %d 涉及日期: %v", userID, getDateKeys(dateMap))

		for date, daySamples := range dateMap {
			log.Printf("📊 处理日期: %s, 样本数: %d", date, len(daySamples))

			// 计算聚合值
			agg := calculateAggregation(userID, date, daySamples)
			log.Printf("📈 聚合结果: HRV=%v, 心率=%v, 步数=%d",
				agg.AvgHrv, agg.AvgHeartRate, agg.TotalSteps)

			// 写入数据库（Upsert）
			if err := upsertDailyAgg(db, agg); err != nil {
				log.Printf("❌ 数据库写入失败: %v", err)
			} else {
				log.Printf("✅ 数据库写入成功: user_id=%d, date=%s", userID, date)
			}

			// 更新 Redis 缓存
			if err := updateCache(rdb, agg); err != nil {
				log.Printf("❌ Redis 缓存失败: %v", err)
			} else {
				log.Printf("✅ Redis 缓存成功: user_id=%d, date=%s", userID, date)
			}
		}
	}
	log.Println("✅ 聚合完成")
}

// getDateKeys 获取日期 map 的键列表
func getDateKeys(m map[string][]models.HealthSample) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// calculateAggregation 计算单日聚合值
func calculateAggregation(userID int64, date string, samples []models.HealthSample) models.DailyAgg {
	var agg models.DailyAgg
	agg.UserID = userID
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Printf("⚠️ 日期解析失败: %s, %v", date, err)
		agg.Date = time.Now()
	} else {
		agg.Date = parsedDate
	}

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
func upsertDailyAgg(db *gorm.DB, agg models.DailyAgg) error {
	// 如果所有值都是 nil/0，跳过插入
	if agg.AvgHrv == nil && agg.AvgHeartRate == nil && agg.TotalSteps == 0 {
		log.Println("⚠️ 聚合数据全为空，跳过插入")
		return nil
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
		log.Printf("❌ upsertDailyAgg 执行失败: %v", err)
		return err
	}
	return nil
}

// updateCache 更新 Redis 缓存
func updateCache(rdb *redis.Client, agg models.DailyAgg) error {
	if rdb == nil {
		return nil
	}

	ctx := context.Background()
	dateStr := agg.Date.Format("2006-01-02")
	cacheKey := "comp:" + string(rune(agg.UserID)) + ":" + dateStr

	jsonData, err := json.Marshal(agg)
	if err != nil {
		log.Printf("❌ JSON 序列化失败: %v", err)
		return err
	}

	if err := rdb.Set(ctx, cacheKey, jsonData, 24*time.Hour).Err(); err != nil {
		log.Printf("❌ Redis Set 失败: %v", err)
		return err
	}
	return nil
}

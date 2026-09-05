package models

import (
	"time"
)

// User 用户表
type User struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"uniqueIndex;type:varchar(64);not null" json:"user_id"` // ✅ 存苹果 sub，保持 string
	Email     *string   `gorm:"type:varchar(128)" json:"email"`
	Name      *string   `gorm:"type:varchar(64)" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// HealthSample 健康原始数据表
type HealthSample struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SampleUUID string    `gorm:"uniqueIndex;type:varchar(64);not null" json:"sample_uuid"`
	UserID     int64     `gorm:"index:idx_user_date;not null" json:"user_id"` // ✅ int64，关联 users.ID
	DataType   string    `gorm:"type:varchar(20);not null" json:"data_type"`
	Value      float64   `gorm:"type:double precision;not null" json:"value"`
	StartDate  time.Time `gorm:"index:idx_user_date;type:timestamptz;not null" json:"start_date"`
	EndDate    time.Time `gorm:"type:timestamptz;not null" json:"end_date"`
	Source     string    `gorm:"type:varchar(20);default:iphone" json:"source"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (HealthSample) TableName() string {
	return "health_samples"
}

// DailyAgg 日聚合表（表盘专用）
type DailyAgg struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64     `gorm:"uniqueIndex:idx_user_date;not null" json:"user_id"` // ✅ int64，关联 users.ID
	Date         time.Time `gorm:"uniqueIndex:idx_user_date;type:date;not null" json:"date"`
	AvgHrv       *float64  `gorm:"type:double precision" json:"avg_hrv"`
	AvgHeartRate *float64  `gorm:"type:double precision" json:"avg_heart_rate"`
	TotalSteps   int64     `gorm:"default:0" json:"total_steps"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (DailyAgg) TableName() string {
	return "daily_agg"
}

// ApnsToken APNs 推送令牌表
type ApnsToken struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"index;not null" json:"user_id"` // ✅ int64，关联 users.ID
	DeviceType string    `gorm:"type:varchar(10);not null" json:"device_type"`
	Token      string    `gorm:"uniqueIndex;type:varchar(128);not null" json:"token"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ApnsToken) TableName() string {
	return "apns_tokens"
}

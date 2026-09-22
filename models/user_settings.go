package models

import (
	"time"
)

// UserSettings 用户设置表
type UserSettings struct {
	ID     int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID int64 `gorm:"uniqueIndex;not null" json:"user_id"` // 关联 users.ID

	// HRV 相关设置
	HrvMin            float64 `gorm:"default:30" json:"hrv_min"`              // HRV 最低阈值
	HrvMax            float64 `gorm:"default:60" json:"hrv_max"`              // HRV 最高阈值
	HrvAlertThreshold float64 `gorm:"default:0.2" json:"hrv_alert_threshold"` // HRV 波动告警阈值（20%）

	// 步数目标
	StepsGoal int `gorm:"default:8000" json:"steps_goal"`

	// 推送开关
	WeatherPushEnabled bool `gorm:"default:true" json:"weather_push_enabled"`
	SleepPushEnabled   bool `gorm:"default:true" json:"sleep_push_enabled"`
	EmotionPushEnabled bool `gorm:"default:true" json:"emotion_push_enabled"`

	// 推送时间（如 "07:00"）
	PushTime string `gorm:"type:varchar(10);default:'07:00'" json:"push_time"`

	// 天气城市
	WeatherCity string `gorm:"type:varchar(64);default:'北京'" json:"weather_city"`

	// 单位制：metric（公制）/ imperial（英制）
	UnitSystem string `gorm:"type:varchar(10);default:'metric'" json:"unit_system"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

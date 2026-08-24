package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"watch-api/config"
	"watch-api/models"
)

func InitDB(cfg config.Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 开发时打印 SQL
	})
	if err != nil {
		log.Fatalf("❌ 连接数据库失败: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ 获取 sql.DB 失败: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移（建表/更新表结构）
	if err := db.AutoMigrate(
		&models.User{},
		&models.HealthSample{},
		&models.DailyAgg{},
		&models.ApnsToken{},
	); err != nil {
		log.Fatalf("❌ 自动迁移失败: %v", err)
	}

	log.Println("✅ PostgreSQL 连接成功，表已就绪")
	return db
}

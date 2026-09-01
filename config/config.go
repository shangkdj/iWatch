package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	// PostgreSQL
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	SSLMode    string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func LoadConfig() Config {
	cfg := Config{
		DBHost:        os.Getenv("DB_HOST"),
		DBPort:        os.Getenv("DB_PORT"),
		DBUser:        os.Getenv("DB_USER"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_NAME"),
		SSLMode:       os.Getenv("SSL_MODE"),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       getEnvInt("REDIS_DB", 0),
	}

	// 检查必须的配置项（非敏感）
	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		log.Fatal("❌ 缺少必要的数据库配置，请检查环境变量 DB_HOST, DB_USER, DB_NAME")
	}
	if cfg.RedisAddr == "" {
		log.Fatal("❌ 缺少 Redis 配置，请检查环境变量 REDIS_ADDR")
	}

	log.Println("✅ 配置加载成功")
	return cfg
}

// getEnvInt 读取整数型环境变量，带默认值（仅用于非敏感数字，如 RedisDB）
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

package config

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
	return Config{
		// PostgreSQL 配置（改成你自己的）
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "Admin%100",
		DBName:     "health_app",
		SSLMode:    "disable",

		// Redis 配置
		RedisAddr:     "localhost:6379",
		RedisPassword: "Admin%100",
		RedisDB:       0,
	}
}

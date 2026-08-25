package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"watch-api/config"
	"watch-api/db"
	"watch-api/handlers"
)

func main() {
	// 1. 加载配置
	cfg := config.LoadConfig()

	// 2. 初始化数据库
	database := db.InitDB(cfg)

	// 3. 初始化 Redis
	rdb := db.InitRedis(cfg)
	_ = rdb // 暂时用不到，占位

	// 4. 延迟关闭连接
	defer func() {
		sqlDB, _ := database.DB()
		if err := sqlDB.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		}
		if err := rdb.Close(); err != nil {
			log.Printf("关闭 Redis 连接失败: %v", err)
		}
		log.Println("🛑 服务已关闭，连接已释放")
	}()

	// 5. 创建 Gin 引擎
	r := gin.Default()

	// 6. 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "200",
			"msg":    "服务运行正常",
		})
	})

	// 7. 测试 GORM 的接口（通过闭包注入 db）
	r.GET("/test-db", handlers.TestDB(database))

	// 8. 启动服务器
	log.Println("🚀 服务启动在 http://127.0.0.1:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("❌ 服务启动失败: %v", err)
	}
}

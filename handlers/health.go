package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// 只做一个测试接口，验证 Gin + GORM 是否正常
func TestDB(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 检查数据库连接是否存活
        sqlDB, err := db.DB()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "获取数据库连接失败: " + err.Error(),
            })
            return
        }

        if err := sqlDB.Ping(); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "数据库 Ping 失败: " + err.Error(),
            })
            return
        }

        // 简单测试查询
        var count int64
        db.Model(&struct{}{}).Raw("SELECT 1").Count(&count)

        c.JSON(http.StatusOK, gin.H{
            "status":  "ok",
            "message": "Gin + GORM 集成成功！",
            "db_ping": "success",
        })
    }
}
package main

import (
	"fmt"
	"project/config"
	"project/models"
	"project/routes"
	"project/utils"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// 设置全局时区为 UTC
	time.Local = time.UTC

	// Initialize database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&models.Shop{}); err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	// Pass config to utils
	utils.InitConfig(cfg)

	// Setup and run server
	r := routes.SetupRouter(db)
	r.Run(":8080")
}

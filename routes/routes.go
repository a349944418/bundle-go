package routes

import (
	"project/config"
	"project/controllers"
	"project/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	r := gin.Default()

	// 加载 HTML 模板
	r.LoadHTMLGlob("templates/*")

	// 公开路由
	r.GET("/install", controllers.InstallHandler(db, cfg))
	r.GET("/callback", controllers.CallbackHandler(db, cfg))

	// 受保护路由组（包括首页和 protected）
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(db))
	{
		protected.GET("/home", controllers.HomeHandler(db))
		protected.GET("/protected", controllers.ProtectedHandler(db))
	}

	return r
}

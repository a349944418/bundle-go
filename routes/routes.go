package routes

import (
	"path/filepath"
	"project/config"
	"project/controllers"
	"project/middleware"
	"strings"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func createRender() multitemplate.Renderer {
	render := multitemplate.NewRenderer()
	pages, err := filepath.Glob("templates/*.html")
	if err != nil {
		panic(err.Error())
	}

	// For each page, generate a template including the layouts.
	for _, page := range pages {
		render.AddFromFiles(strings.TrimSuffix(filepath.Base(page), ".html"), "templates/layout/base.html", page)
	}

	return render
}

func SetupRouter(db *gorm.DB) *gin.Engine {
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	r := gin.Default()

	// 使用 multitemplate
	r.HTMLRender = createRender()

	// 公开路由
	r.GET("/install", controllers.InstallHandler(db, cfg))
	r.GET("/callback", controllers.CallbackHandler(db, cfg))

	// 受保护路由组（包括首页和 protected）
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(db))
	{
		protected.GET("/", controllers.HomeHandler(db))
		protected.GET("/protected", controllers.ProtectedHandler(db))
	}

	return r
}

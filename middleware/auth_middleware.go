package middleware

import (
	"fmt"
	"net/http"
	"project/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware 检查令牌有效性
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		shop := c.Query("handle")
		if shop == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少店铺参数"})
			c.Abort()
			return
		}

		token, err := services.GetToken(db, shop)
		if err != nil {
			fmt.Printf("GetToken error: %v\n", err)
			c.Redirect(http.StatusFound, "/install?handle="+shop)
			c.Abort()
			return
		}

		if services.IsTokenExpired(token) || !services.ValidateToken(shop, token.AccessToken) {
			fmt.Printf("StatusUnauthorized令牌已过期或无效，请重新授权: %v\n", err)
			c.Redirect(http.StatusFound, "/install?handle="+shop)
			c.Abort()
			return
		}

		// 令牌有效，继续处理请求
		c.Next()
	}
}

package middleware

import (
	"net/http"
	"project/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware 检查令牌有效性
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		shop := c.Query("shop")
		if shop == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少店铺参数"})
			c.Abort()
			return
		}

		token, err := services.GetToken(db, shop)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到令牌，请重新安装应用"})
			c.Redirect(http.StatusFound, "/install?shop="+shop)
			c.Abort()
			return
		}

		if services.IsTokenExpired(token) || !services.ValidateToken(shop, token.AccessToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌已过期或无效，请重新授权"})
			c.Redirect(http.StatusFound, "/install?shop="+shop)
			c.Abort()
			return
		}

		// 令牌有效，继续处理请求
		c.Next()
	}
}

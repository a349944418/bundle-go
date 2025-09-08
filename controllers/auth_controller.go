package controllers

import (
	"net/http"
	"project/config"
	"project/services"
	"project/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InstallHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		shop := c.Query("handle")
		if shop == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少店铺参数"})
			return
		}

		authURL := utils.BuildAuthURL(shop)

		c.Redirect(http.StatusFound, authURL)
	}
}

func CallbackHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		handle := c.Query("handle")
		code := c.Query("code")
		hmacStr := c.Query("sign")

		if handle == "" || code == "" || hmacStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
			return
		}

		query := c.Request.URL.RawQuery
		if !utils.VerifyHMAC(strings.Replace(query, "sign="+hmacStr, "", 1), hmacStr) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "HMAC 验证失败"})
			return
		}

		token, err := services.ExchangeCodeForToken(handle, code, cfg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取令牌失败: " + err.Error()})
			return
		}

		if err := services.StoreShopToken(db, handle, token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "存储令牌失败: " + err.Error()})
			return
		}

		// 授权成功，重定向到首页
		c.Redirect(http.StatusFound, "/?handle="+handle)
	}
}

func HomeHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		shop := c.Query("handle")
		c.HTML(http.StatusOK, "home.html", gin.H{
			"ShopDomain": shop,
		})
	}
}

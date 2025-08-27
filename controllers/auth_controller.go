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
		shop := c.Query("shop")
		if shop == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少店铺参数"})
			return
		}

		nonce, err := utils.GenerateNonce()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 nonce 失败"})
			return
		}

		utils.StoreNonce(nonce)

		authURL := utils.BuildAuthURL(shop, nonce)
		c.Redirect(http.StatusFound, authURL)
	}
}

func CallbackHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		shop := c.Query("shop")
		code := c.Query("code")
		state := c.Query("state")
		hmacStr := c.Query("hmac")

		if shop == "" || code == "" || state == "" || hmacStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
			return
		}

		query := c.Request.URL.RawQuery
		if !utils.VerifyHMAC(strings.Replace(query, "hmac="+hmacStr, "", 1), hmacStr) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "HMAC 验证失败"})
			return
		}

		if !utils.VerifyNonce(state) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 state"})
			return
		}

		token, err := services.ExchangeCodeForToken(shop, code, cfg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取令牌失败: " + err.Error()})
			return
		}

		if err := services.StoreShopToken(db, shop, token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "存储令牌失败: " + err.Error()})
			return
		}

		// 授权成功，重定向到首页
		c.Redirect(http.StatusFound, "/home?shop="+shop)
	}
}

func ProtectedHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "受保护的内容 - 令牌有效"})
	}
}

func HomeHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		shop := c.Query("shop")
		c.HTML(http.StatusOK, "home.html", gin.H{
			"ShopDomain": shop,
		})
	}
}

package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"project/config"
	"project/models"
	"project/utils"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type tokenResponse struct {
    Code     int    `json:"code"`
    I18nCode string `json:"i18nCode"`
    Message  string `json:"message"`
    Data     struct {
        AccessToken string `json:"accessToken"`
        ExpireTime  string `json:"expireTime"`
        Scope       string `json:"scope"`
		RefreshToken string `json:"refreshToken"`
    } `json:"data"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	ExpiresAt    time.Time
}

func ExchangeCodeForToken(shop string, code string, cfg *config.Config) (Token, error) {
	tokenURL := fmt.Sprintf("https://%s.myshopline.com/admin/oauth/token/create", shop)
	payload := `{"code":"`+code+`"}`

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(payload))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return Token{}, err
	}

	millis := time.Now().UnixMilli()
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("timestamp", strconv.FormatInt(millis, 10))
	req.Header.Set("appKey", cfg.Shopline.APIKey)
	req.Header.Set("sign", utils.HmacSha256(payload+strconv.FormatInt(millis, 10)))

	// 创建 HTTP 客户端并发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return Token{}, errors.New("Token request failed: " + string(body))
	}

	var respData tokenResponse
	// 解析响应体
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return Token{}, err
	}

	token := Token{
        AccessToken: respData.Data.AccessToken,
        Scope:       respData.Data.Scope,
        RefreshToken: respData.Data.RefreshToken,
    }

	expireTimeLocal := utils.CovertToLocalTime(respData.Data.ExpireTime)
	if expireTimeLocal.After(time.Now()) {
		token.ExpiresAt = expireTimeLocal.UTC()
	}
	fmt.Printf("Exchanged code for token: %+v\n", token)
	return token, nil
}

func StoreShopToken(db *gorm.DB, shop string, token Token) error {
	shopRecord := models.Shop{
		ShopDomain:   shop,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Scope:        token.Scope,
		ExpiresAt:    token.ExpiresAt,
		InstalledAt:  time.Now(),
	}
	return db.Where("shop_domain = ?", shop).FirstOrCreate(&shopRecord).Error
}

func GetToken(db *gorm.DB, shop string) (Token, error) {
	var shopRecord models.Shop
	if err := db.Where("shop_domain = ?", shop).First(&shopRecord).Error; err != nil {
		return Token{}, err
	}
	return Token{
		AccessToken:  shopRecord.AccessToken,
		RefreshToken: shopRecord.RefreshToken,
		Scope:        shopRecord.Scope,
		ExpiresAt:    shopRecord.ExpiresAt,
	}, nil
}

func IsTokenExpired(token Token) bool {
	if token.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(token.ExpiresAt)
}

func ValidateToken(shop string, accessToken string) bool {
	apiURL := fmt.Sprintf("https://%s.myshopline.com/admin/openapi/v20251201/products/products.json", shop)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "bearer "+accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	fmt.Printf("Validating token with request: %+v\n", resp)
	return resp.StatusCode == http.StatusOK
}

package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"project/config"
	"project/models"
	"time"

	"gorm.io/gorm"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	ExpiresAt    time.Time
}

func ExchangeCodeForToken(shop string, code string, cfg *config.Config) (Token, error) {
	tokenURL := fmt.Sprintf("https://%s/admin/oauth/access_token", shop)
	data := url.Values{
		"client_id":     {cfg.Shopify.APIKey},
		"client_secret": {cfg.Shopify.APISecret},
		"code":          {code},
	}
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return Token{}, errors.New("Token request failed: " + string(body))
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return Token{}, err
	}

	if token.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}

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
	return db.Save(&shopRecord).Error
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
	apiURL := fmt.Sprintf("https://%s/admin/api/2023-10/shop.json", shop)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("X-Shopify-Access-Token", accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"project/config"
	"time"

	"net/url"
)

var cfg *config.Config

func InitConfig(config *config.Config) {
	cfg = config
}

func VerifyHMAC(query string, hmacStr string) bool {
	params, err := url.ParseQuery(query)
	if err != nil {
		return false
	}
	params.Del("sign")
	encoded := params.Encode()
	h := hmac.New(sha256.New, []byte(cfg.Shopline.APISecret))
	h.Write([]byte(encoded))
	expected := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(hmacStr))
}

func HmacSha256(source string) string {
    if source != "" {
        h := hmac.New(sha256.New, []byte(cfg.Shopline.APISecret))
        h.Write([]byte(source))
        return hex.EncodeToString(h.Sum(nil))
    }
    return ""
}

func BuildAuthURL(shop string) string {
	return fmt.Sprintf("https://%s.myshopline.com/admin/oauth-web/#/oauth/authorize?appKey=%s&scope=%s&responseTypec=code&redirectUri=%s",
		shop, cfg.Shopline.APIKey, cfg.Shopline.Scopes, url.QueryEscape(cfg.Shopline.RedirectURI))
}

func CovertToLocalTime(covertTime string) time.Time {
	layout := "2006-01-02T15:04:05.000Z07:00"
	t, err := time.Parse(layout, covertTime)
	if err != nil {
		fmt.Println("解析失败:", err)
		return time.Time{}
	}
	
	localTime := t.In(time.Local)
	return localTime
}


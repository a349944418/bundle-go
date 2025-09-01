package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"project/config"

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

func BuildAuthURL(shop string) string {
	return fmt.Sprintf("https://%s.myshopline.com/admin/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s",
		shop, cfg.Shopline.APIKey, cfg.Shopline.Scopes, url.QueryEscape(cfg.Shopline.RedirectURI))
}

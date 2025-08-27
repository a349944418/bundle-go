package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"project/config"
	"sync"
	"time"

	"net/url"
)

var cfg *config.Config

// NonceStore
var nonceStore = struct {
	sync.Mutex
	data map[string]time.Time
}{data: make(map[string]time.Time)}

func InitConfig(config *config.Config) {
	cfg = config
}

func GenerateNonce() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	return base64.StdEncoding.EncodeToString(b), err
}

func StoreNonce(nonce string) {
	nonceStore.Lock()
	nonceStore.data[nonce] = time.Now()
	nonceStore.Unlock()
}

func VerifyNonce(state string) bool {
	nonceStore.Lock()
	_, exists := nonceStore.data[state]
	delete(nonceStore.data, state)
	nonceStore.Unlock()
	return exists
}

func VerifyHMAC(query string, hmacStr string) bool {
	params, err := url.ParseQuery(query)
	if err != nil {
		return false
	}
	params.Del("hmac")
	encoded := params.Encode()
	h := hmac.New(sha256.New, []byte(cfg.Shopify.APISecret))
	h.Write([]byte(encoded))
	expected := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(hmacStr))
}

func BuildAuthURL(shop, nonce string) string {
	return fmt.Sprintf("https://%s/admin/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s&state=%s",
		shop, cfg.Shopify.APIKey, cfg.Shopify.Scopes, url.QueryEscape(cfg.Shopify.RedirectURI), nonce)
}

func CleanupNonces() {
	for {
		time.Sleep(5 * time.Minute)
		nonceStore.Lock()
		now := time.Now()
		for k, v := range nonceStore.data {
			if now.Sub(v) > 10*time.Minute {
				delete(nonceStore.data, k)
			}
		}
		nonceStore.Unlock()
	}
}

package futures

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type FutureClient struct {
	APIKey    string
	SecretKey string
	BaseURL   string
}

func NewClient(apiKey, secretKey string) *FutureClient {
	return &FutureClient{
		APIKey:    apiKey,
		SecretKey: secretKey,
		BaseURL:   "https://fapi.binance.com",
	}
}

func (f *FutureClient) sign(params string) string {
	h := hmac.New(sha256.New, []byte(f.SecretKey))
	h.Write([]byte(params))
	return hex.EncodeToString(h.Sum(nil))
}

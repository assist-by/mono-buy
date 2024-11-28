package futures

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type CandleData struct {
	OpenTime                 int64  `json:"openTime"`
	Open                     string `json:"open"`
	High                     string `json:"high"`
	Low                      string `json:"low"`
	Close                    string `json:"close"`
	Volume                   string `json:"volume"`
	CloseTime                int64  `json:"closeTime"`
	QuoteAssetVolume         string `json:"quoteAssetVolume"`
	NumberOfTrades           int    `json:"numberOfTrades"`
	TakerBuyBaseAssetVolume  string `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume string `json:"takerBuyQuoteAssetVolume"`
}

type Balance struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"free"`
	Locked float64 `json:"locked"`
	Total  float64 `json:"total"`
}

type FutureClient struct {
	APIKey     string
	SecretKey  string
	BaseURL    string
	MaxRetries int
	RetryDelay time.Duration
}

func NewClient(apiKey, secretKey string) *FutureClient {
	return &FutureClient{
		APIKey:     apiKey,
		SecretKey:  secretKey,
		BaseURL:    "https://fapi.binance.com",
		MaxRetries: 5,
		RetryDelay: 5 * time.Second,
	}
}

func (f *FutureClient) sign(params string) string {
	h := hmac.New(sha256.New, []byte(f.SecretKey))
	h.Write([]byte(params))
	return hex.EncodeToString(h.Sum(nil))
}

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

type AccountBalance struct {
	Asset              string  `json:"asset"`
	CrossWalletBalance float64 `json:"crossWalletBalance,string"`
	CrossUnPnl         float64 `json:"crossUnPnl,string"`
	AvailableBalance   float64 `json:"availableBalance,string"`
	WalletBalance      float64 `json:"walletBalance,string"`
}

type Balance struct {
	Free   float64
	Locked float64
}

type FutureClient struct {
	APIKey           string
	SecretKey        string
	BaseURL          string
	ServerTimeOffset int64
	MaxRetries       int
	RetryDelay       time.Duration
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

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type WalletBalance struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"free"`
	Locked float64 `json:"locked"`
	Total  float64 `json:"total"`
}

func fetchWalletBalance(apiKey, secretKey string) (map[string]WalletBalance, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/account", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	timestamp := fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
	req.Header.Set("X-MBX-APIKEY", apiKey)

	q := req.URL.Query()
	q.Add("timestamp", timestamp)
	signature := hmacSha256([]byte(q.Encode()), []byte(secretKey))
	q.Add("signature", signature)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request fialed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response faield: %w", err)
	}

	var account struct {
		Balances []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		} `json:"balances"`
	}

	if err = json.Unmarshal(body, &account); err != nil {
		return nil, fmt.Errorf("parsing response failed: %w", err)
	}

	balances := make(map[string]WalletBalance)

	for _, balance := range account.Balances {
		free, err := strconv.ParseFloat(balance.Free, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing free balance fialed for %s: %w", balance.Asset, err)
		}

		locked, err := strconv.ParseFloat(balance.Locked, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing locked balance failed for %s: %w", balance.Asset, err)
		}

		total := free + locked

		if total > 0 {
			balances[balance.Asset] = WalletBalance{
				Asset:  balance.Asset,
				Free:   free,
				Locked: locked,
				Total:  total,
			}
		}
	}

	return balances, nil
}

// func getAssetsBalance(balances map[string]WalletBalance, asset string) (WalletBalance, error) {
// 	if balance, exists := balances[asset]; exists {
// 		return balance, nil
// 	}
// 	return WalletBalance{}, fmt.Errorf("asset %s not found in wallet", asset)

// }

func hmacSha256(message, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}

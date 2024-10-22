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

func fetchWalletBalance(apiKey, secretKey string) (float64, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/account", nil)
	if err != nil {
		return 0, err
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
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var account struct {
		Balances []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		} `json:"balances"`
	}

	err = json.Unmarshal(body, &account)
	if err != nil {
		return 0, err
	}

	for _, balance := range account.Balances {
		if balance.Asset == "BTC" {
			btcBalance, err := strconv.ParseFloat(balance.Free, 64)
			if err != nil {
				return 0, err
			}
			return btcBalance, nil
		}
	}

	return 0, fmt.Errorf("BTC balance not found")
}

func hmacSha256(message, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}

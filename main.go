package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	binanceKlineAPI = "https://api.binance.com/api/v3/klines"
	candleLimit     = 1
	fetchInterval   = 1 * time.Minute
)

type CandleData struct {
	OpenTime  int64
	Open      string
	High      string
	Low       string
	Close     string
	Volume    string
	CloseTime int64
}

func fetchBTCCandleData(url string) ([]CandleData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var klines [][]interface{}
	err = json.Unmarshal(body, &klines)
	if err != nil {
		return nil, err
	}

	candles := make([]CandleData, len(klines))
	for i, kline := range klines {
		candles[i] = CandleData{
			OpenTime:  int64(kline[0].(float64)),
			Open:      kline[1].(string),
			High:      kline[2].(string),
			Low:       kline[3].(string),
			Close:     kline[4].(string),
			Volume:    kline[5].(string),
			CloseTime: int64(kline[6].(float64)),
		}
	}

	return candles, nil
}

func waitUntilNextMinute() {
	now := time.Now()
	next := now.Add(time.Minute).Truncate(time.Minute)
	time.Sleep(next.Sub(now))
}

func main() {
	for {
		waitUntilNextMinute()

		url := fmt.Sprintf("%s?symbol=BTCUSDT&interval=1m&limit=%d", binanceKlineAPI, candleLimit)

		candles, err := fetchBTCCandleData(url)
		if err != nil {
			log.Printf("Error fetching candle data: %v\n", err)
			continue
		}

		if len(candles) > 0 {
			candle := candles[0]
			currentTime := time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("Bitcoin Price Data (as of %s):\n", currentTime)
			fmt.Printf("Close: %s USDT\n", candle.Close)
			fmt.Printf("24h High: %s USDT\n", candle.High)
			fmt.Printf("24h Low: %s USDT\n", candle.Low)
			fmt.Printf("24h Volume: %s BTC\n", candle.Volume)
			fmt.Println("------------------------")
		}
	}
}

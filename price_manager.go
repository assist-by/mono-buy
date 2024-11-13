package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/assist-by/mono-buy/lib"
)

var intervalMap = map[time.Duration]string{
	time.Minute:         "1m",
	15 * time.Minute:    "15m",
	30 * time.Minute:    "30m",
	time.Hour:           "1h",
	2 * time.Hour:       "2h",
	4 * time.Hour:       "4h",
	6 * time.Hour:       "6h",
	8 * time.Hour:       "8h",
	12 * time.Hour:      "12h",
	24 * time.Hour:      "1d",
	7 * 24 * time.Hour:  "1w",
	30 * 24 * time.Hour: "1M",
}

func nextIntervalStart(now time.Time, interval time.Duration) time.Time {
	return now.Truncate(interval).Add(interval)
}

func getIntervalString(interval time.Duration) string {
	if str, exists := intervalMap[interval]; exists {
		return str
	}
	return "15m" // 기본값
}

func fetchBTCCandleData(url string) ([]lib.CandleData, error) {
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

	candles := make([]lib.CandleData, len(klines))
	for i, kline := range klines {
		candles[i] = lib.CandleData{
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

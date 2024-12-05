package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"sort"
// 	"strconv"
// 	"strings"

// 	lib "github.com/assist-by/libStruct"
// )

// const (
// 	binanceFuturesExchangeInfo = "https://fapi.binance.com/fapi/v1/exchangeInfo"
// 	binanceFutures24hr         = "https://fapi.binance.com/fapi/v1/ticker/24hr"
// )

// func getTopVolumeSymbols(limit int) ([]string, error) {
// 	resp, err := http.Get(binanceFutures24hr)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch 24hr data: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read resp body: %v", err)
// 	}

// 	var tickers []struct {
// 		Symbol      string `json:"symbol"`
// 		QuoteVolume string `json:"quoteVolume"`
// 	}

// 	if err := json.Unmarshal(body, &tickers); err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal resp : %v", err)
// 	}

// 	// USDT 페어만 필터링
// 	var coins []lib.Coin
// 	for _, ticker := range tickers {
// 		if !strings.HasSuffix(ticker.Symbol, "USDT") {
// 			continue
// 		}

// 		quoteVolume, err := strconv.ParseFloat(ticker.QuoteVolume, 64)
// 		if err != nil {
// 			continue
// 		}

// 		coins = append(coins, lib.Coin{
// 			Symbol:      ticker.Symbol,
// 			QuoteVolume: quoteVolume,
// 		})
// 	}

// 	// 거래량 기준 내림차순 정렬
// 	sort.Slice(coins, func(i, j int) bool {
// 		return coins[i].QuoteVolume > coins[j].QuoteVolume
// 	})

// 	// 상위 N개 반환
// 	result := make([]string, 0, limit)
// 	for i := 0; i < limit && i < len(coins); i++ {
// 		result = append(result, coins[i].Symbol)
// 	}

// 	return result, nil
// }

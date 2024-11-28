package futures

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (f *FutureClient) GetTopVolumeSymbols(n int) ([]string, error) {
	endpoint := "/fapi/v1/ticker/24hr"

	req, err := http.NewRequest("GET", f.BaseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	var tickers []struct {
		Symbol      string  `json:"symbol"`
		QuoteVolume float64 `json:"quoteVolume,string"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if err := json.Unmarshal(body, &tickers); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	// USDT 마진 선물만 필터링하고 바로 정렬
	var filteredTickers []struct {
		Symbol      string  `json:"symbol"`
		QuoteVolume float64 `json:"quoteVolume,string"`
	}

	for _, ticker := range tickers {
		if strings.HasSuffix(ticker.Symbol, "USDT") {
			filteredTickers = append(filteredTickers, ticker)
		}
	}

	// Sort by volume
	sort.Slice(filteredTickers, func(i, j int) bool {
		return filteredTickers[i].QuoteVolume > filteredTickers[j].QuoteVolume
	})

	// 상위 20개 거래량 시각화
	maxBars := 20
	if len(filteredTickers) > 0 {
		maxVolume := filteredTickers[0].QuoteVolume
		fmt.Println("\n=== Top Volume Symbols ===")
		for i := 0; i < maxBars && i < len(filteredTickers); i++ {
			ticker := filteredTickers[i]
			barLength := int((ticker.QuoteVolume / maxVolume) * 50) // 50은 최대 막대 길이
			bar := strings.Repeat("=", barLength)
			fmt.Printf("%-12s %15.2f ||%s\n", ticker.Symbol, ticker.QuoteVolume, bar)
		}
		fmt.Println("========================")
	}

	symbols := make([]string, 0, n)
	for i := 0; i < n && i < len(filteredTickers); i++ {
		symbols = append(symbols, filteredTickers[i].Symbol)
	}

	return symbols, nil
}

func (f *FutureClient) GetKlineData(symbol string, interval string, limit int) ([]CandleData, error) {
	params := url.Values{}
	params.Add("symbol", symbol)
	params.Add("interval", interval)
	params.Add("limit", strconv.Itoa(limit))

	endpoint := "/fapi/v1/klines"
	req, err := http.NewRequest("GET", f.BaseURL+endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	var rawCandles [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawCandles); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	candles := make([]CandleData, len(rawCandles))
	for i, raw := range rawCandles {
		candles[i] = CandleData{
			OpenTime:                 int64(raw[0].(float64)),
			Open:                     raw[1].(string),
			High:                     raw[2].(string),
			Low:                      raw[3].(string),
			Close:                    raw[4].(string),
			Volume:                   raw[5].(string),
			CloseTime:                int64(raw[6].(float64)),
			QuoteAssetVolume:         raw[7].(string),
			NumberOfTrades:           int(raw[8].(float64)),
			TakerBuyBaseAssetVolume:  raw[9].(string),
			TakerBuyQuoteAssetVolume: raw[10].(string),
		}
	}

	return candles, nil
}

func (f *FutureClient) GetWalletBalance() (map[string]Balance, error) {
	timestamp := time.Now().UnixMilli()
	params := url.Values{}
	params.Add("timestamp", strconv.FormatInt(timestamp, 10))

	signature := f.sign(params.Encode())
	params.Add("signature", signature)

	endpoint := "/fapi/v2/balance"
	req, err := http.NewRequest("GET", f.BaseURL+endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", f.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	var balances []struct {
		Asset  string  `json:"asset"`
		Free   float64 `json:"free,string"`
		Locked float64 `json:"locked,string"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&balances); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	result := make(map[string]Balance)
	for _, b := range balances {
		result[b.Asset] = Balance{
			Asset:  b.Asset,
			Free:   b.Free,
			Locked: b.Locked,
			Total:  b.Free + b.Locked,
		}
	}

	return result, nil

}

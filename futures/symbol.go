package futures

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type SymbolInfo struct {
	Symbol              string `json:"symbol"`
	PricePrecision      int    `json:"pricePrecision"`
	QuantityPrecision   int    `json:"quantityPrecision"`
	BaseAssetPrecision  int    `json:"baseAssetPrecision"`
	QuoteAssetPrecision int    `json:"quotePrecision"`
	MinNotional         float64
	StepSize            float64
}

// 심볼 정보 조회
func (f *FutureClient) GetSymbolInfo(symbol string) (*SymbolInfo, error) {
	endpoint := "/fapi/v1/exchangeInfo"

	req, err := http.NewRequest("GET", f.BaseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var result struct {
		Symbols []struct {
			Symbol            string `json:"symbol"`
			PricePrecision    int    `json:"pricePrecision"`
			QuantityPrecision int    `json:"quantityPrecision"`
			Filters           []struct {
				FilterType  string `json:"filterType"`
				MinNotional string `json:"notional,omitempty"`
				StepSize    string `json:"stepSize,omitempty"`
			} `json:"filters"`
		} `json:"symbols"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	for _, s := range result.Symbols {
		if s.Symbol == symbol {
			info := &SymbolInfo{
				Symbol:            s.Symbol,
				PricePrecision:    s.PricePrecision,
				QuantityPrecision: s.QuantityPrecision,
			}

			// Parse filters
			for _, filter := range s.Filters {
				switch filter.FilterType {
				case "MIN_NOTIONAL":
					info.MinNotional, _ = strconv.ParseFloat(filter.MinNotional, 64)
				case "LOT_SIZE":
					info.StepSize, _ = strconv.ParseFloat(filter.StepSize, 64)
				}
			}

			return info, nil
		}
	}

	return nil, fmt.Errorf("symbol not found: %s", symbol)
}

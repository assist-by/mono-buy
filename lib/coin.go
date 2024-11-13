package lib

type Coin struct {
	Symbol      string  `json:"symbol"`
	Volume      float64 `json:"volume"`
	QuoteVolume float64 `json:"quoteVolume"`
}

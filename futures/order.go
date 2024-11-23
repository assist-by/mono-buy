package futures

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type OrderSide string
type PositionSide string

const (
	BUY  OrderSide = "BUY"
	SELL OrderSide = "SELL"

	LONG  PositionSide = "LONG"
	SHORT PositionSide = "SHORT"
)

type OrderRequest struct {
	Symbol       string
	Side         OrderSide
	PositionSide PositionSide
	Type         string
	Quantity     float64
	Price        float64
	StopPrice    float64
	TakeProfit   float64
	StopLoss     float64
	TimeInForce  string
}

func (f *FutureClient) PlaceOrder(order OrderRequest) error {
	timestamp := time.Now().UnixMilli()
	params := url.Values{}
	params.Add("symbol", order.Symbol)
	params.Add("side", string(order.Side))
	params.Add("positionSide", string(order.PositionSide))
	params.Add("type", "MARKET")
	params.Add("quantity", strconv.FormatFloat(order.Quantity, 'f', -1, 64))
	params.Add("timestamp", strconv.FormatInt(timestamp, 10))

	signature := f.sign(params.Encode())
	params.Add("signature", signature)

	req, err := http.NewRequest("POST", f.BaseURL+"/fapi/v1/order", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", f.APIKey)
	req.URL.RawQuery = params.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}

	defer resp.Body.Close()
}

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

	if order.TakeProfit > 0 {
		tpParams := url.Values{}
		tpParams.Add("symbol", order.Symbol)
		tpParams.Add("side", string(getOppositeOrderSide(order.Side)))
		tpParams.Add("type", "TAKE_PROFIT_MARKET")
		tpParams.Add("stopPrice", strconv.FormatFloat(order.TakeProfit, 'f', -1, 64))
		tpParams.Add("quantity", strconv.FormatFloat(order.Quantity, 'f', -1, 64))
		tpParams.Add("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

		signature = f.sign(tpParams.Encode())
		tpParams.Add("signature", signature)

		if err := f.placeStopOrder(tpParams); err != nil {
			return fmt.Errorf("placing take profit: %w", err)
		}
	}

	if order.StopLoss > 0 {
		slParams := url.Values{}
		slParams.Add("symbol", order.Symbol)
		slParams.Add("side", string(getOppositeOrderSide(order.Side)))
		slParams.Add("type", "STOP_MARKET")
		slParams.Add("stopPrice", strconv.FormatFloat(order.StopLoss, 'f', -1, 64))
		slParams.Add("quantity", strconv.FormatFloat(order.Quantity, 'f', -1, 64))
		slParams.Add("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

		signature = f.sign(slParams.Encode())
		slParams.Add("signature", signature)

		if err := f.placeStopOrder(slParams); err != nil {
			return fmt.Errorf("placing stop loss: %w", err)
		}
	}

	return nil
}

func (f *FutureClient) placeStopOrder(params url.Values) error {
	req, err := http.NewRequest("POST", f.BaseURL+"/fapi/v1/order", nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-MBX-APIKEY", f.APIKey)
	req.URL.RawQuery = params.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

func getOppositeOrderSide(side OrderSide) OrderSide {
	if side == BUY {
		return SELL
	}

	return BUY
}

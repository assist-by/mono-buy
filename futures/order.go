package futures

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
	// 테스트를 위해 주석 처리

	// timestamp := time.Now().UnixMilli()
	// params := url.Values{}
	// params.Add("symbol", order.Symbol)
	// params.Add("side", string(order.Side))
	// params.Add("positionSide", string(order.PositionSide))
	// params.Add("type", "MARKET")
	// params.Add("quantity", strconv.FormatFloat(order.Quantity, 'f', -1, 64))
	// params.Add("timestamp", strconv.FormatInt(timestamp, 10))

	// signature := f.sign(params.Encode())
	// params.Add("signature", signature)

	// req, err := http.NewRequest("POST", f.BaseURL+"/fapi/v1/order", nil)
	// if err != nil {
	// 	return fmt.Errorf("creating request: %w", err)
	// }

	// req.Header.Set("X-MBX-APIKEY", f.APIKey)
	// req.URL.RawQuery = params.Encode()

	// resp, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	return fmt.Errorf("sending request: %w", err)
	// }

	// defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	body, _ := io.ReadAll(resp.Body)
	// 	return fmt.Errorf("order failed: %s", string(body))
	// }

	// if order.TakeProfit > 0 {
	// 	tpParams := url.Values{}
	// 	tpParams.Add("symbol", order.Symbol)
	// 	tpParams.Add("side", string(getOppositeOrderSide(order.Side)))
	// 	tpParams.Add("type", "TAKE_PROFIT_MARKET")
	// 	tpParams.Add("stopPrice", strconv.FormatFloat(order.TakeProfit, 'f', -1, 64))
	// 	tpParams.Add("quantity", strconv.FormatFloat(order.Quantity, 'f', -1, 64))
	// 	tpParams.Add("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	// 	signature = f.sign(tpParams.Encode())
	// 	tpParams.Add("signature", signature)

	// 	if err := f.placeStopOrder(tpParams); err != nil {
	// 		return fmt.Errorf("placing take profit: %w", err)
	// 	}
	// }

	// if order.StopLoss > 0 {
	// 	slParams := url.Values{}
	// 	slParams.Add("symbol", order.Symbol)
	// 	slParams.Add("side", string(getOppositeOrderSide(order.Side)))
	// 	slParams.Add("type", "STOP_MARKET")
	// 	slParams.Add("stopPrice", strconv.FormatFloat(order.StopLoss, 'f', -1, 64))
	// 	slParams.Add("quantity", strconv.FormatFloat(order.Quantity, 'f', -1, 64))
	// 	slParams.Add("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	// 	signature = f.sign(slParams.Encode())
	// 	slParams.Add("signature", signature)

	// 	if err := f.placeStopOrder(slParams); err != nil {
	// 		return fmt.Errorf("placing stop loss: %w", err)
	// 	}
	// }

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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stop order failed: %s", string(body))
	}
	return nil
}

func getOppositeOrderSide(side OrderSide) OrderSide {
	if side == BUY {
		return SELL
	}

	return BUY
}

func (f *FutureClient) SetLeverage(symbol string, leverage int) error {
	// 서버 시간 가져오기
	serverTime, err := f.GetServerTime()
	if err != nil {
		return fmt.Errorf("getting server time: %w", err)
	}

	params := url.Values{}
	params.Add("symbol", symbol)
	params.Add("leverage", strconv.Itoa(leverage))
	params.Add("timestamp", strconv.FormatInt(serverTime, 10))
	params.Add("recvWindow", "10000")

	signature := f.sign(params.Encode())
	params.Add("signature", signature)

	endpoint := "/fapi/v1/leverage"

	req, err := http.NewRequest("POST", f.BaseURL+endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", f.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("setting leverage failed: %s", string(body))
	}

	return nil
}

// 포지션 모드 설정 (Hedge Mode / One-way Mode)

func (f *FutureClient) SetPositionMode(hedgeMode bool) error {
	serverTime, err := f.GetServerTime()
	if err != nil {
		return fmt.Errorf("getting server time: %w", err)
	}

	params := url.Values{}
	params.Add("dualSidePosition", strconv.FormatBool(hedgeMode))
	params.Add("timestamp", strconv.FormatInt(serverTime, 10))
	params.Add("recvWindow", "10000")

	signature := f.sign(params.Encode())
	params.Add("signature", signature)

	endpoint := "/fapi/v1/positionSide/dual"
	req, err := http.NewRequest("POST", f.BaseURL+endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", f.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		// Binance 에러 응답 파싱
		var binanceErr BinanceError
		if err := json.Unmarshal(body, &binanceErr); err != nil {
			return fmt.Errorf("parsing error response: %w", err)
		}

		// 이미 설정되어 있는 경우
		if binanceErr.Code == ERROR_NO_NEED_TO_CHANGE_POSITION {
			return nil
		}

		return fmt.Errorf("setting position mode failed: %s", string(body))
	}

	return nil
}

// // 수량을 stepSize에 맞게 내림하는 유틸리티 함수
// func FloorToStepSize(quantity float64, stepSize float64) float64 {
// 	precision := getPrecisionFromStepSize(stepSize)
// 	factor := float64(10 * precision)
// 	return float64(int(quantity*factor)) / factor
// }

func getPrecisionFromStepSize(stepSize float64) int {
	precision := 0
	for stepSize < 1 {
		stepSize *= 10
		precision++
	}
	return precision
}

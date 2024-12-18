package futures

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (f *FutureClient) GetServerTime() (int64, error) {
	resp, err := http.Get(f.BaseURL + "/fapi/v1/time")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		ServerTime int64 `json:"serverTime"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.ServerTime, nil
}

// 서버 시간 동기화 함수
func (f *FutureClient) SyncServerTime() error {
	resp, err := http.Get(f.BaseURL + "/fapi/v1/time")
	if err != nil {
		return fmt.Errorf("getting server time: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ServerTime int64 `json:"serverTime"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	f.ServerTimeOffset = result.ServerTime - time.Now().UnixMilli()
	return nil
}

// 타임스탬프 생성 시 오프셋 적용
func (f *FutureClient) GetTimestamp() int64 {
	return time.Now().UnixMilli() + f.ServerTimeOffset
}

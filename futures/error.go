package futures

const (
	ERROR_NO_NEED_TO_CHANGE_POSITION = -4059
)

type BinanceError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

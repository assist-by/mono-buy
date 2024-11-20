package lib

type TechnicalIndicators struct {
	EMA200       float64
	ParabolicSAR float64
	MACDLine     float64
	SignalLine   float64
}

// MACD 크로스 체크를 위한 구조체
type MACDCross struct {
	CurrentMACDLine   float64
	CurrentSignalLine float64
	PrevMACDLine      float64
	PrevSignalLine    float64
}

package lib

type SignalConditions struct {
	Long  SignalDetail
	Short SignalDetail
}

type SignalDetail struct {
	EMA200Condition       bool
	ParabolicSARCondition bool
	MACDCondition         bool
	EMA200Value           float64
	EMA200Diff            float64
	ParabolicSARValue     float64
	ParabolicSARDiff      float64
	MACDMACDLine          float64
	MACDSignalLine        float64
	MACDHistogram         float64
}

type SignalResult struct {
	Symbol     string
	Signal     string
	Timestamp  int64
	Price      float64
	StopLoss   float64
	TakeProfit float64
	Conditions SignalConditions
}

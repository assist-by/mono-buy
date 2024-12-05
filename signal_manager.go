package main

import (
	"fmt"
	"strconv"

	"github.com/assist-by/abmodule/calculate"
	lib "github.com/assist-by/libStruct"
	future "github.com/assist-by/mono-buy/futures"
)

// func processCandles(candles []future.CandleData, now time.Time) (lastCompleteCandle future.CandleData, err error) {
// 	if len(candles) < 2 {
// 		return future.CandleData{}, fmt.Errorf("insufficient candles data")
// 	}

// 	// 마지막 캔들의 종료 시간을 확인
// 	lastCandle := candles[len(candles)-1]
// 	lastCandleClose := time.Unix(lastCandle.CloseTime/1000, 0)

// 	// 현재 시간이 마지막 캔들의 종료 시간보다 크거나 같으면
// 	// 마지막 캔들이 완료된 캔들임
// 	if now.Equal(lastCandleClose) || now.After(lastCandleClose) {
// 		return lastCandle, nil
// 	}

// 	// 그렇지 않으면 마지막에서 두 번째 캔들이 완료된 마지막 캔들
// 	return candles[len(candles)-2], nil
// }

// 매수 신호 생성 함수
func generateSignal(candles []future.CandleData, indicators lib.TechnicalIndicators) (lib.SignalType, lib.SignalConditions, float64, float64) {
	if len(candles) < 2 { // 최소 2개의 캔들 필요
		// 캔들조회 에러
		return lib.SIGNAL_NO_SIGANL, lib.SignalConditions{}, 0.0, 0.0
	}

	lastPrice, _ := strconv.ParseFloat(candles[len(candles)-1].Close, 64)
	lastHigh, _ := strconv.ParseFloat(candles[len(candles)-1].High, 64)
	lastLow, _ := strconv.ParseFloat(candles[len(candles)-1].Low, 64)

	prevPrices := make([]float64, len(candles)-1)
	for i := 0; i < len(candles)-1; i++ {
		price, _ := strconv.ParseFloat(candles[i].Close, 64)
		prevPrices[i] = price
	}

	prevMACDLine, prevSignalLine := calculate.CalculateMACD(prevPrices)

	macdCross := lib.MACDCross{
		CurrentMACDLine:   indicators.MACDLine,
		CurrentSignalLine: indicators.SignalLine,
		PrevMACDLine:      prevMACDLine,
		PrevSignalLine:    prevSignalLine,
	}

	upCross := macdCross.PrevMACDLine < macdCross.PrevSignalLine && macdCross.CurrentMACDLine > macdCross.CurrentSignalLine

	downCross := macdCross.PrevMACDLine > macdCross.PrevSignalLine && macdCross.CurrentMACDLine < macdCross.CurrentSignalLine

	conditions := lib.SignalConditions{
		Long: lib.SignalDetail{
			EMA200Condition:       lastPrice > indicators.EMA200,
			ParabolicSARCondition: indicators.ParabolicSAR < lastLow,
			MACDCondition:         upCross, // 크로스 조건으로 변경
			EMA200Value:           indicators.EMA200,
			EMA200Diff:            lastPrice - indicators.EMA200,
			ParabolicSARValue:     indicators.ParabolicSAR,
			ParabolicSARDiff:      lastLow - indicators.ParabolicSAR,
			MACDHistogram:         indicators.MACDLine - indicators.SignalLine,
			MACDMACDLine:          indicators.MACDLine,
			MACDSignalLine:        indicators.SignalLine,
		},
		Short: lib.SignalDetail{
			EMA200Condition:       lastPrice < indicators.EMA200,
			ParabolicSARCondition: indicators.ParabolicSAR > lastHigh,
			MACDCondition:         downCross, // 크로스 조건으로 변경
			EMA200Value:           indicators.EMA200,
			EMA200Diff:            lastPrice - indicators.EMA200,
			ParabolicSARValue:     indicators.ParabolicSAR,
			ParabolicSARDiff:      indicators.ParabolicSAR - lastHigh,
			MACDHistogram:         indicators.MACDLine - indicators.SignalLine,
			MACDMACDLine:          indicators.MACDLine,
			MACDSignalLine:        indicators.SignalLine,
		},
	}

	// 최대 손절 거리
	const maxStopLossDistance = 0.007 // 0.7%
	var stopLoss, takeProfit float64

	if conditions.Long.EMA200Condition && conditions.Long.ParabolicSARCondition && conditions.Long.MACDCondition {
		stopLoss = indicators.ParabolicSAR
		// Long 포지션의 경우
		if lastPrice-stopLoss > lastPrice*maxStopLossDistance {
			stopLoss = lastPrice * (1 - maxStopLossDistance)
		}
		takeProfit = lastPrice + (lastPrice - stopLoss)
		return lib.SIGNAL_LONG, conditions, stopLoss, takeProfit
	} else if conditions.Short.EMA200Condition && conditions.Short.ParabolicSARCondition && conditions.Short.MACDCondition {
		stopLoss = indicators.ParabolicSAR
		// Short 포지션의 경우
		if stopLoss-lastPrice > lastPrice*maxStopLossDistance {
			stopLoss = lastPrice * (1 + maxStopLossDistance)
		}
		takeProfit = lastPrice - (stopLoss - lastPrice)
		return lib.SIGNAL_SHORT, conditions, stopLoss, takeProfit
	}

	return lib.SIGNAL_NO_SIGANL, conditions, 0.0, 0.0
}

// 보조지표값 계산 함수
func calculateIndicators(candles []future.CandleData) (lib.TechnicalIndicators, error) {
	if len(candles) < 300 {
		return lib.TechnicalIndicators{}, fmt.Errorf("insufficient data: need at least 300 candles, got %d", len(candles))
	}

	prices := make([]float64, len(candles))
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))

	for i, candle := range candles {
		price, err := strconv.ParseFloat(candle.Close, 64)
		if err != nil {
			return lib.TechnicalIndicators{}, fmt.Errorf("error parsing close price: %v", err)
		}
		prices[i] = price

		high, err := strconv.ParseFloat(candle.High, 64)
		if err != nil {
			return lib.TechnicalIndicators{}, fmt.Errorf("error parsing high price: %v", err)
		}
		highs[i] = high

		low, err := strconv.ParseFloat(candle.Low, 64)
		if err != nil {
			return lib.TechnicalIndicators{}, fmt.Errorf("error parsing low price: %v", err)
		}
		lows[i] = low
	}

	ema200 := calculate.CalculateEMA(prices, 200)
	macdLine, signalLine := calculate.CalculateMACD(prices)
	parabolicSAR := calculate.CalculateParabolicSAR(highs, lows)

	return lib.TechnicalIndicators{
		EMA200:       ema200,
		ParabolicSAR: parabolicSAR,
		MACDLine:     macdLine,
		SignalLine:   signalLine,
	}, nil
}

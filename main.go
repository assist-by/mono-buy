package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	calculate "github.com/assist-by/abmodule/calculate"
	notification "github.com/assist-by/abmodule/notification"
	lib "github.com/assist-by/libStruct"
	signalType "github.com/assist-by/libStruct/enums/signalType"
	"github.com/joho/godotenv"
)

const (
	binanceKlineAPI = "https://api.binance.com/api/v3/klines"
	maxRetries      = 5
	retryDelay      = 5 * time.Second
	candleLimit     = 300
	fetchInterval   = 1 * time.Minute
)

var (
	isRunning         bool
	discordWebhookURL string
	runningMutex      sync.Mutex
	serviceCtx        context.Context
	serviceCtxCancel  context.CancelFunc
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	discordWebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")

	serviceCtx, serviceCtxCancel = context.WithCancel(context.Background())
}

func fetchBTCCandleData(url string) ([]lib.CandleData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var klines [][]interface{}
	err = json.Unmarshal(body, &klines)
	if err != nil {
		return nil, err
	}

	candles := make([]lib.CandleData, len(klines))
	for i, kline := range klines {
		candles[i] = lib.CandleData{
			OpenTime:  int64(kline[0].(float64)),
			Open:      kline[1].(string),
			High:      kline[2].(string),
			Low:       kline[3].(string),
			Close:     kline[4].(string),
			Volume:    kline[5].(string),
			CloseTime: int64(kline[6].(float64)),
		}
	}

	return candles, nil
}

func calculateIndicators(candles []lib.CandleData) (lib.TechnicalIndicators, error) {
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

func generateSignal(candles []lib.CandleData, indicators lib.TechnicalIndicators) (string, lib.SignalConditions, float64, float64) {
	lastPrice, _ := strconv.ParseFloat(candles[len(candles)-1].Close, 64)
	lastHigh, _ := strconv.ParseFloat(candles[len(candles)-1].High, 64)
	lastLow, _ := strconv.ParseFloat(candles[len(candles)-1].Low, 64)

	conditions := lib.SignalConditions{
		Long: lib.SignalDetail{
			EMA200Condition:       lastPrice > indicators.EMA200,
			ParabolicSARCondition: indicators.ParabolicSAR < lastLow,
			MACDCondition:         indicators.MACDLine > indicators.SignalLine,
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
			MACDCondition:         indicators.MACDLine < indicators.SignalLine,
			EMA200Value:           indicators.EMA200,
			EMA200Diff:            lastPrice - indicators.EMA200,
			ParabolicSARValue:     indicators.ParabolicSAR,
			ParabolicSARDiff:      indicators.ParabolicSAR - lastHigh,
			MACDHistogram:         indicators.MACDLine - indicators.SignalLine,
			MACDMACDLine:          indicators.MACDLine,
			MACDSignalLine:        indicators.SignalLine,
		},
	}

	var stopLoss, takeProfit float64

	if conditions.Long.EMA200Condition && conditions.Long.ParabolicSARCondition && conditions.Long.MACDCondition {
		stopLoss = indicators.ParabolicSAR
		takeProfit = lastPrice + (lastPrice - stopLoss)
		return signalType.Long.String(), conditions, stopLoss, takeProfit
	} else if conditions.Short.EMA200Condition && conditions.Short.ParabolicSARCondition && conditions.Short.MACDCondition {
		stopLoss = indicators.ParabolicSAR
		takeProfit = lastPrice - (stopLoss - lastPrice)
		return signalType.Short.String(), conditions, stopLoss, takeProfit
	}
	return signalType.No_Signal.String(), conditions, 0.0, 0.0
}

func nextIntervalStart(now time.Time, interval time.Duration) time.Time {
	return now.Truncate(interval).Add(interval)
}

func getIntervalString(interval time.Duration) string {
	switch interval {
	case 1 * time.Minute:
		return "1m"
	case 15 * time.Minute:
		return "15m"
	case 1 * time.Hour:
		return "1h"
	default:
		return "15m"
	}
}

func processSignal(signalResult lib.SignalResult) error {
	log.Printf("Processing signal: %+v", signalResult)

	discordColor := notification.GetColorForDiscord(signalResult.Signal)

	title := fmt.Sprintf("New Signal: %s", signalResult.Signal)
	description := generateDescription(signalResult)

	discordEmbed := notification.Embed{
		Title:       title,
		Description: description,
		Color:       discordColor,
	}
	if err := notification.SendDiscordAlert(discordEmbed, discordWebhookURL); err != nil {
		log.Printf("Error sending Discord alert: %v", err)
		return err
	}

	log.Println("Notifications sent successfully")
	return nil
}

func generateDescription(signalResult lib.SignalResult) string {
	koreaLocation, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		log.Printf("Error loading Asia/Seoul timezone: %v", err)
		koreaLocation = time.UTC
	}
	timestamp := time.Unix(signalResult.Timestamp/1000, 0).In(koreaLocation).Format("2006-01-02 15:04:05 MST")

	description := fmt.Sprintf("Signal: %s for BTCUSDT at %s\n\n", signalResult.Signal, timestamp)
	description += fmt.Sprintf("Price : %.3f\n", signalResult.Price)

	if signalResult.Signal != signalType.No_Signal.String() {
		description += fmt.Sprintf("Stoploss : %.3f, Takeprofit: %.3f\n\n", signalResult.StopLoss, signalResult.TakeProfie)
	}

	description += "=======[LONG]=======\n"
	description += fmt.Sprintf("[EMA200] : %v \n", signalResult.Conditions.Long.EMA200Condition)
	description += fmt.Sprintf("EMA200: %.3f, Diff: %.3f\n\n", signalResult.Conditions.Long.EMA200Value, signalResult.Conditions.Long.EMA200Diff)

	description += fmt.Sprintf("[MACD] : %v \n", signalResult.Conditions.Long.MACDCondition)
	description += fmt.Sprintf("Now MACD Line: %.3f, Now Signal Line: %.3f, Now Histogram: %.3f\n", signalResult.Conditions.Long.MACDMACDLine, signalResult.Conditions.Long.MACDSignalLine, signalResult.Conditions.Long.MACDHistogram)

	description += fmt.Sprintf("[Parabolic SAR] : %v \n", signalResult.Conditions.Long.ParabolicSARCondition)
	description += fmt.Sprintf("ParabolicSAR: %.3f, Diff: %.3f\n", signalResult.Conditions.Long.ParabolicSARValue, signalResult.Conditions.Long.ParabolicSARDiff)
	description += "=====================\n\n"

	description += "=======[SHORT]=======\n"
	description += fmt.Sprintf("[EMA200] : %v \n", signalResult.Conditions.Short.EMA200Condition)
	description += fmt.Sprintf("EMA200: %.3f, Diff: %.3f\n\n", signalResult.Conditions.Short.EMA200Value, signalResult.Conditions.Short.EMA200Diff)

	description += fmt.Sprintf("[MACD] : %v \n", signalResult.Conditions.Short.MACDCondition)
	description += fmt.Sprintf("MACD Line: %.3f, Signal Line: %.3f, Histogram: %.3f\n\n", signalResult.Conditions.Short.MACDMACDLine, signalResult.Conditions.Short.MACDSignalLine, signalResult.Conditions.Short.MACDHistogram)

	description += fmt.Sprintf("[Parabolic SAR] : %v \n", signalResult.Conditions.Short.ParabolicSARCondition)
	description += fmt.Sprintf("ParabolicSAR: %.3f, Diff: %.3f\n", signalResult.Conditions.Short.ParabolicSARValue, signalResult.Conditions.Short.ParabolicSARDiff)
	description += "=====================\n"

	return description
}

func startService(ctx context.Context) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for {
		now := time.Now()
		nextFetch := nextIntervalStart(now, fetchInterval)
		sleepDuration := nextFetch.Sub(now)

		log.Printf("Waiting for %v until next fetch at %v\n", sleepDuration.Round(time.Second), nextFetch.Format("2006-01-02 15:04:05"))

		select {
		case <-time.After(sleepDuration):
			url := fmt.Sprintf("%s?symbol=BTCUSDT&interval=%s&limit=%d", binanceKlineAPI, getIntervalString(fetchInterval), candleLimit)

			candles, err := fetchBTCCandleData(url)
			if err != nil {
				log.Printf("Error fetching candle data: %v\n", err)
				continue
			}

			if len(candles) == candleLimit {
				indicators, err := calculateIndicators(candles)
				if err != nil {
					log.Printf("Error calculating indicators: %v\n", err)
					continue
				}

				signalType, conditions, stopLoss, takeProfit := generateSignal(candles, indicators)
				lastCandle := candles[len(candles)-1]
				price, err := strconv.ParseFloat(lastCandle.Close, 64)
				if err != nil {
					log.Printf("Error convert price to float: %v\n", err)
					continue
				}

				signalResult := lib.SignalResult{
					Signal:     signalType,
					Timestamp:  lastCandle.CloseTime,
					Price:      price,
					Conditions: conditions,
					StopLoss:   stopLoss,
					TakeProfie: takeProfit,
				}

				if err := processSignal(signalResult); err != nil {
					log.Printf("Error processing signal: %v", err)
				}
			} else {
				log.Printf("Insufficient data: got %d candles, expected %d\n", len(candles), candleLimit)
			}

		case <-signals:
			log.Println("Interrupt received, shutting down...")
			return

		case <-ctx.Done():
			log.Println("Context cancelled, shutting down...")
			return
		}
	}
}

func main() {

	log.Println("Starting BTC Signal Generator with Notifications...")

	runningMutex.Lock()
	isRunning = true
	runningMutex.Unlock()

	startService(serviceCtx)

	log.Println("BTC Signal Generator with Notifications stopped")
}

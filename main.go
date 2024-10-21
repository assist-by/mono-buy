package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	lib "github.com/assist-by/libStruct"
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

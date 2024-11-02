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
	apikey            string
	secretkey         string
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
	apikey = os.Getenv("API_KEY")
	secretkey = os.Getenv("SECRET_KEY")

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
			balances, err := fetchWalletBalance(apikey, secretkey)
			if err != nil {
				log.Printf("❌ Error fetching wallet balances: %v\n", err)
			} else {
				log.Printf("=== 현재 지갑 상태 ===")
				if let(balances) == 0 {
					log.Printf("⚠️ 잔액이 있는 자산이 없습니다.")
				} else {
					for asset, balance := range balances {
						log.Printf("🏦 %s 보유량: %.8f (가용: %.8f, 잠금: %.8f)\n",
							asset, balance.Total, balance.Free, balance.Locked)
					}
				}
			}
			log.Printf("-------------------------------------------")

			// 가격 데이터 조회
			url := fmt.Sprintf("%s?symbol=BTCUSDT&interval=%s&limit=%d", binanceKlineAPI, getIntervalString(fetchInterval), candleLimit)

			candles, err := fetchBTCCandleData(url)
			if err != nil {
				log.Printf("Error fetching candle data: %v\n", err)
				continue
			}

			if len(candles) == candleLimit {
				// 현재 BTC 가격 계산
				currentPrice, _ := strconv.ParseFloat(candles[len(candles)-1].Close, 64)
				log.Printf("💰 현재 BTC 가격: $%.2f\n", currentPrice)

				// 보조지표 계산
				indicators, err := calculateIndicators(candles)
				if err != nil {
					log.Printf("❌ Error calculating indicators: %v\n", err)
					continue
				}

				// 매수매도 신호 계산
				signalType, conditions, stopLoss, takeProfit := generateSignal(candles, indicators)
				lastCandle := candles[len(candles)-1]
				price, err := strconv.ParseFloat(lastCandle.Close, 64)
				if err != nil {
					log.Printf("❌ Error convert price to float: %v\n", err)
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

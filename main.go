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
	/// TODO: 선물거래 가격을 조회하자
	binanceKlineAPI = "https://api.binance.com/api/v3/klines"
	maxRetries      = 5
	retryDelay      = 5 * time.Second
	candleLimit     = 1000
	// fetchInterval   = 1 * time.Minute
)

var (
	apikey            string
	secretkey         string
	isRunning         bool
	discordWebhookURL string
	fetchInterval     time.Duration
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
	fetchInterval, err = time.ParseDuration(os.Getenv("FETCH_INTERVAL"))
	if err != nil {
		log.Fatalf("Invalid fetch interval: %v", err)
	}
	serviceCtx, serviceCtxCancel = context.WithCancel(context.Background())
}

func NewCoinTracker(symbol string) *lib.CoinTracker {
	return &lib.CoinTracker{
		Symbol: symbol,
	}
}
func startService(ctx context.Context) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	trackers := make(map[string]*lib.CoinTracker)

	for {
		now := time.Now()
		nextCheck := now.Truncate(fetchInterval).Add(fetchInterval)
		sleepDuration := nextCheck.Sub(now)

		log.Printf("Waiting for %v until next fetch at %v\n", sleepDuration.Round(time.Second), nextCheck.Format("2006-01-02 15:04:05"))

		select {
		case <-time.After(sleepDuration):
			topSymbols, err := getTopVolumeSymbols(3)
			if err != nil {
				log.Printf("❌ Error fetching top volume symbols: %v\n", err)
				continue
			}

			log.Printf("🔍 현재 추적 중인 상위 코인: %v\n", topSymbols)

			balances, err := fetchWalletBalance(apikey, secretkey)
			if err != nil {
				log.Printf("❌ Error fetching wallet balances: %v\n", err)
			} else {
				log.Printf("=== 현재 지갑 상태 ===")
				if len(balances) == 0 {
					log.Printf("⚠️ 잔액이 있는 자산이 없습니다.")
				} else {
					for asset, balance := range balances {
						log.Printf("🏦 %s 보유량: %.8f (가용: %.8f, 잠금: %.8f)\n",
							asset, balance.Total, balance.Free, balance.Locked)
					}
				}
			}
			log.Printf("-------------------------------------------")

			for _, symbol := range topSymbols {
				if _, exists := trackers[symbol]; !exists {
					trackers[symbol] = NewCoinTracker(symbol)
				}

				url := fmt.Sprintf("%s?symbol=%s&interval=%s&limit=%d",
					binanceKlineAPI,
					symbol,
					getIntervalString(fetchInterval),
					candleLimit)

				candles, err := fetchCandleData(url)
				if err != nil {
					log.Printf("❌ Error fetching candle data for %s: %v\n", symbol, err)
					continue
				}

				if len(candles) < 2 {
					log.Printf("Insufficient data for %s: got %d candles\n", symbol, len(candles))
					continue
				}

				lastCompletedCandle := candles[len(candles)-2]
				indicators, err := calculateIndicators(candles)
				if err != nil {
					log.Printf("❌ Error calculating indicators for %s: %v\n", symbol, err)
					continue
				}

				signalType, conditions, stopLoss, takeProfit := generateSignal(candles, indicators)

				if signalType != trackers[symbol].LastSignal || lastCompletedCandle.CloseTime != trackers[symbol].LastSignalTime {
					price, err := strconv.ParseFloat(lastCompletedCandle.Close, 64)
					if err != nil {
						log.Printf("❌ Error converting price for %s: %v\n", symbol, err)
						continue
					}

					signalResult := lib.SignalResult{
						Symbol:     symbol,
						Signal:     signalType,
						Timestamp:  lastCompletedCandle.CloseTime,
						Price:      price,
						Conditions: conditions,
						StopLoss:   stopLoss,
						TakeProfit: takeProfit,
					}

					if err := processSignal(signalResult); err != nil {
						log.Printf("Error processing signal for %s: %v", symbol, err)
					}

					trackers[symbol].LastSignal = signalType
					trackers[symbol].LastSignalTime = lastCompletedCandle.CloseTime
				}
			}

			for symbol := range trackers {
				found := false
				for _, topSymbol := range topSymbols {
					if symbol == topSymbol {
						found = true
						break
					}
				}
				if !found {
					delete(trackers, symbol)
				}
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

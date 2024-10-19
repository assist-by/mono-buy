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
	"sync"
	"time"

	lib "github.com/assist-by/libStruct"
	"github.com/gin-gonic/gin"
)

const (
	binanceKlineAPI = "https://api.binance.com/api/v3/klines"
	maxRetries      = 5
	retryDelay      = 5 * time.Second
	candleLimit     = 300
	fetchInterval   = 15 * time.Minute
)

var (
	host             string
	port             string
	isRunning        bool
	runningMutex     sync.Mutex
	serviceCtx       context.Context
	serviceCtxCancel context.CancelFunc
)

func init() {
	host = os.Getenv("HOST")
	if host == "" {
		host = "abprice"
	}
	port = os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}
	serviceCtx, serviceCtxCancel = context.WithCancel(context.Background())
}

// 캔들 데이터 패치
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

func printCandles(candles []lib.CandleData) {
	for _, candle := range candles {
		fmt.Printf("OpenTime: %d, Open: %s, High: %s, Low: %s, Close: %s, Volume: %s, CloseTime: %d\n",
			candle.OpenTime, candle.Open, candle.High, candle.Low, candle.Close, candle.Volume, candle.CloseTime)
	}
}

func utcToLocal(utcTime time.Time) time.Time {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		log.Printf("Error loading location: %v\n", err)
		return utcTime
	}
	return utcTime.In(loc)
}

// 다음 fetch 시간 구하는 함수
func nextIntervalStart(now time.Time, interval time.Duration) time.Time {
	return now.Truncate(interval).Add(interval)
}

// 시간 반복에 따른 url에 넣을 String 반환 함수
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

// 서비스 시작 함수
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
			url := fmt.Sprintf("%s?symbol=BTCUSDT&interval=%s&limit=%d", binanceKlineAPI, getIntervalString(sleepDuration), candleLimit)

			candles, err := fetchBTCCandleData(url)
			if err != nil {
				log.Printf("Error fetching candle data: %v\n", err)
				continue
			}

			printCandles(candles)
			log.Printf("Successfully printed %d candle data\n", len(candles))
			if len(candles) > 0 {
				firstCandle := candles[0]
				lastCandle := candles[len(candles)-1]
				firstTime := utcToLocal(time.Unix(firstCandle.OpenTime/1000, 0))
				lastTime := utcToLocal(time.Unix(lastCandle.CloseTime/1000, 0))
				log.Printf("Data range (Local Time): %v to %v\n",
					firstTime.Format("2006-01-02 15:04:05"),
					lastTime.Format("2006-01-02 15:04:05"))
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

// POST
// start service를 /start API로 받아서 실행하는 함수
func startHandler(c *gin.Context) {
	runningMutex.Lock()
	defer runningMutex.Unlock()

	if isRunning {
		c.JSON(http.StatusOK, gin.H{"message": "abprice is already running"})
		return
	}

	isRunning = true
	go startService(serviceCtx)
	c.JSON(http.StatusOK, gin.H{"message": "abprice started successfully"})
}

// POST
// stop service를 /stop API로 받아서 실행하는 함수
func stopHandler(c *gin.Context) {
	runningMutex.Lock()
	defer runningMutex.Unlock()

	if !isRunning {
		c.JSON(http.StatusOK, gin.H{"message": "abprice is not running"})
		return
	}

	serviceCtxCancel() // 서비스 컨텍스트 취소
	isRunning = false
	c.JSON(http.StatusOK, gin.H{"message": "abprice stopped successfully"})
}

// GET
// status를 /status API로 받아서 실행하는 함수
func statusHandler(c *gin.Context) {
	runningMutex.Lock()
	defer runningMutex.Unlock()

	status := "stopped"
	if isRunning {
		status = "running"
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}

func main() {
	router := gin.Default()
	router.POST("/start", startHandler)
	router.POST("/stop", stopHandler)
	router.GET("/status", statusHandler)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server : %v", err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	<-signals
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

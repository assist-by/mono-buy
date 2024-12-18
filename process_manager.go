package main

import (
	"fmt"
	"log"

	lib "github.com/assist-by/libStruct"
	"github.com/assist-by/mono-buy/discord"
	"github.com/assist-by/mono-buy/futures"
)

func processSignal(signalResult lib.SignalResult) error {
	log.Printf("Processing signal for %s...", signalResult.Symbol)

	// 주문 전송
	if err := sendNotification(signalResult); err != nil {
		log.Printf("❌ Error sending notification for %s: %v", signalResult.Symbol, err)
		return fmt.Errorf("sending notification for %s: %w", signalResult.Symbol, err)
	}

	// // 시그널이 없으면 처리하지 않음
	// if signalResult.Signal == lib.SIGNAL_NO_SIGANL {
	// 	log.Printf("No signal for %s, skipping", signalResult.Symbol)
	// 	return nil
	// }

	// 주문 전송
	if err := sendOrder(signalResult); err != nil {
		log.Printf("❌ Error sending order for %s: %v", signalResult.Symbol, err)
		return fmt.Errorf("sending order for %s: %w", signalResult.Symbol, err)
	}

	log.Printf("✅ Successfully processed signal for %s", signalResult.Symbol)
	return nil
}

// send notification
func sendNotification(signalResult lib.SignalResult) error {
	discordClient := discord.NewClient(discordWebhookURL)
	log.Printf("Processing signal: %+v", signalResult)

	// discord embedding
	discordEmbed := generateDiscordEmbed(signalResult)

	// discord로 알림 보내기
	if err := discordClient.Send(discordEmbed); err != nil {
		log.Printf("Error sending Discord alert: %v", err)
		return err
	}

	log.Println("Notifications sent successfully")

	return nil
}
func sendOrder(signalResult lib.SignalResult) error {
	log.Printf("Starting sendOrder for %s", signalResult.Symbol)

	// send buy api
	client := futures.NewClient(apikey, secretkey)
	discordClient := discord.NewClient(discordWebhookTradeURL)

	// 1. Hedge 모드 설정
	log.Printf("Setting hedge mode for %s", signalResult.Symbol)
	if err := client.SetPositionMode(true); err != nil {
		log.Printf("❌ Hedge mode error for %s: %v", signalResult.Symbol, err)
		return fmt.Errorf("setting hedge mode: %w", err)
	}

	// 2. 심볼 정보 조회
	log.Printf("Getting symbol info for %s", signalResult.Symbol)
	symbolInfo, err := client.GetSymbolInfo(signalResult.Symbol)
	if err != nil {
		log.Printf("❌ Symbol info error for %s: %v", signalResult.Symbol, err)
		return fmt.Errorf("getting symbol info: %w", err)
	}

	// 3. 레버리지 설정
	log.Printf("Setting leverage for %s", signalResult.Symbol)
	if err := client.SetLeverage(signalResult.Symbol, 20); err != nil {
		log.Printf("❌ Leverage error for %s: %v", signalResult.Symbol, err)
		return fmt.Errorf("setting leverage: %w", err)
	}

	// USDT 잔고 조회
	log.Printf("Getting wallet balance for %s", signalResult.Symbol)
	balances, err := client.GetWalletBalance()
	if err != nil {
		log.Printf("❌ Balance error for %s: %v", signalResult.Symbol, err)
		return fmt.Errorf("getting wallet balance: %w", err)
	}

	log.Printf("=== Processing %s ===", signalResult.Symbol)
	usdtBalance := balances["USDT"]
	positionSize := usdtBalance.Free / signalResult.Price
	positionSize = futures.FloorToStepSize(positionSize, symbolInfo.StepSize)

	// 정밀도 로깅 추가
	log.Printf("Raw Position Size before floor: %.8f", positionSize)
	log.Printf("Step Size for %s: %.8f", signalResult.Symbol, symbolInfo.StepSize)
	log.Printf("USDT Balance: %.2f for %s", usdtBalance.Free, signalResult.Symbol)
	log.Printf("Position Size: %.8f for %s", positionSize, signalResult.Symbol)
	log.Printf("Notional Value: %.2f for %s", positionSize*signalResult.Price, signalResult.Symbol)
	log.Printf("Min Notional: %.2f for %s", symbolInfo.MinNotional, signalResult.Symbol)

	// 최소 주문 금액 체크
	if positionSize*signalResult.Price < symbolInfo.MinNotional {
		log.Printf("❗ Order size too small for %s", signalResult.Symbol)
		err := fmt.Errorf("order size too small. minimum notional: %v", symbolInfo.MinNotional)
		if discordClient != nil {
			log.Printf("Sending notification for small order size for %s", signalResult.Symbol)
			if notifyErr := discordClient.SendTradeNotification(signalResult, positionSize, err); notifyErr != nil {
				log.Printf("❌ Failed to send Discord notification for %s: %v", signalResult.Symbol, notifyErr)
			}
		}
		return err
	}

	log.Printf("Passed minimum order check for %s", signalResult.Symbol)

	// 주문 생성
	var order futures.OrderRequest
	switch signalResult.Signal {
	case lib.SIGNAL_LONG:
		order = futures.OrderRequest{
			Symbol:       signalResult.Symbol,
			Side:         futures.BUY,
			PositionSide: futures.LONG,
			Type:         "MARKET",
			Quantity:     positionSize,
			StopLoss:     signalResult.StopLoss,
			TakeProfit:   signalResult.TakeProfit,
		}
		log.Printf("🚀 Opening LONG position for %s at %.2f (TP: %.2f, SL: %.2f)",
			signalResult.Symbol, signalResult.Price, signalResult.TakeProfit, signalResult.StopLoss)

	case lib.SIGNAL_SHORT:
		order = futures.OrderRequest{
			Symbol:       signalResult.Symbol,
			Side:         futures.SELL,
			PositionSide: futures.SHORT,
			Type:         "MARKET",
			Quantity:     positionSize,
			StopLoss:     signalResult.StopLoss,
			TakeProfit:   signalResult.TakeProfit,
		}
		log.Printf("🔻 Opening SHORT position for %s at %.2f (TP: %.2f, SL: %.2f)",
			signalResult.Symbol, signalResult.Price, signalResult.TakeProfit, signalResult.StopLoss)

	case lib.SIGNAL_NO_SIGANL:
		return nil
	}

	// 주문 실행
	if err := func() error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in placing order: %v", r)
				if discordClient != nil {
					discordClient.SendTradeNotification(signalResult, positionSize, fmt.Errorf("order placement panic: %v", r))
				}
			}
		}()

		if err := client.PlaceOrder(order); err != nil {
			log.Printf("Error placing order: %v", err)
			if discordClient != nil {
				discordClient.SendTradeNotification(signalResult, positionSize, err)
			}
			return fmt.Errorf("placing order: %w", err)
		}
		return nil
	}(); err != nil {
		return err
	}

	// 성공 알림 전송
	if discordClient != nil {
		log.Printf("Sending success notification")
		if err := func() error {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic in sending success notification: %v", r)
				}
			}()
			discordClient.SendTradeNotification(signalResult, positionSize, nil)
			return nil
		}(); err != nil {
			log.Printf("Error sending success notification: %v", err)
		}
	}

	return nil
}

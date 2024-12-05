package main

import (
	"fmt"
	"log"

	lib "github.com/assist-by/libStruct"
	"github.com/assist-by/mono-buy/discord"
	"github.com/assist-by/mono-buy/futures"
)

func processSignal(signalResult lib.SignalResult) error {
	// send notification
	sendNotification(signalResult)
	sendOrder(signalResult)
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

// func requestFuture(signalResult lib.SignalResult) error{
// 	clinet := futures.NewClient(apikey,secretkey)

// 	if signalResult.Signal == "Long"
// }

func sendOrder(signalResult lib.SignalResult) error {

	// send buy api
	client := futures.NewClient(apikey, secretkey)
	discordClient := discord.NewClient(discordWebhookTradeURL)

	// 1. Hedge 모드 설정
	if err := client.SetPositionMode(true); err != nil {
		return fmt.Errorf("setting hedge mode: %w", err)
	}

	// 2. 심볼 정보 조회
	symbolInfo, err := client.GetSymbolInfo(signalResult.Symbol)
	if err != nil {
		return fmt.Errorf("getting symbol info: %w", err)
	}

	// 3. 레버리지 설정 (예: 20배)
	if err := client.SetLeverage(signalResult.Symbol, 20); err != nil {
		return fmt.Errorf("setting leverage: %w", err)
	}

	// USDT 잔고 조회
	balances, err := client.GetWalletBalance()
	if err != nil {
		return fmt.Errorf("getting wallet balance: %w", err)
	}

	usdtBalance := balances["USDT"]
	if usdtBalance.Free <= 0 {
		return fmt.Errorf("insufficient USDT balance")
	}

	// 포지션 크기 계산 및 stepSize에 맞게 조정
	positionSize := usdtBalance.Free / signalResult.Price
	positionSize = futures.FloorToStepSize(positionSize, symbolInfo.StepSize)

	// 최소 주문 금액 체크
	if positionSize*signalResult.Price < symbolInfo.MinNotional {
		err := fmt.Errorf("order size too small. minimum notional: %v", symbolInfo.MinNotional)
		discordClient.SendTradeNotification(signalResult, positionSize, err)
		return err
	}

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

	/// 주문하기
	if err := client.PlaceOrder(order); err != nil {
		discordClient.SendTradeNotification(signalResult, positionSize, err)
		return fmt.Errorf("placing order: %w", err)
	}
	discordClient.SendTradeNotification(signalResult, positionSize, nil)
	return nil
}

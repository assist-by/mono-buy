package main

import (
	"fmt"
	"log"
	"time"

	notification "github.com/assist-by/abmodule/notification"
	lib "github.com/assist-by/libStruct"
	signalType "github.com/assist-by/libStruct/enums/signalType"
)

// 로그 description 생성 함수
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

// notification 보내는 함수
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

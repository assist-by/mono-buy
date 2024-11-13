package main

import (
	"fmt"
	"log"
	"time"

	"github.com/assist-by/abmodule/notification"
	"github.com/assist-by/mono-buy/lib"
	signalType "github.com/assist-by/mono-buy/lib/enums/signalType"
)

// processSignal과 generateDiscordEmbed 함수
func generateDiscordEmbed(signalResult lib.SignalResult) notification.Embed {
	koreaLocation, _ := time.LoadLocation("Asia/Seoul")
	timestamp := time.Unix(signalResult.Timestamp/1000, 0).In(koreaLocation)

	var signalEmoji string
	switch signalResult.Signal {
	case signalType.Long.String():
		signalEmoji = "🚀 LONG"
	case signalType.Short.String():
		signalEmoji = "🔻 SHORT"
	default:
		signalEmoji = "⏺️ NO SIGNAL"
	}

	mainDescription := fmt.Sprintf("**시간**: %s\n**현재가**: $%.2f\n",
		timestamp.Format("2006-01-02 15:04:05 KST"),
		signalResult.Price)

	if signalResult.Signal != signalType.No_Signal.String() {
		// 수익률 계산
		stopLossPercent := (signalResult.StopLoss - signalResult.Price) / signalResult.Price * 100
		takeProfitPercent := (signalResult.TakeProfit - signalResult.Price) / signalResult.Price * 100

		// Short 포지션일 경우 수익률 부호를 반대로
		if signalResult.Signal == signalType.Short.String() {
			stopLossPercent = -stopLossPercent
			takeProfitPercent = -takeProfitPercent
		}

		mainDescription += fmt.Sprintf("**스탑로스**: $%.2f (%.2f%%)\n**목표가**: $%.2f (%.2f%%)\n",
			signalResult.StopLoss,
			stopLossPercent,
			signalResult.TakeProfit,
			takeProfitPercent)

	}

	return notification.Embed{
		Title:       fmt.Sprintf("%s BTC/USDT", signalEmoji),
		Description: mainDescription,
		Fields: []notification.EmbedField{
			{
				Name: "📈 LONG",
				Value: fmt.Sprintf("```diff\n%s\n%s\n%s```\n```\n[EMA200]: %.2f (차이: %.2f)\n[MACD Line]: %.2f\n[Signal Line]: %.2f\n[Histogram]: %.2f\n[SAR]: %.2f (차이: %.2f)```",
					formatConditionWithSymbol(signalResult.Conditions.Long.EMA200Condition, "EMA200"),
					formatConditionWithSymbol(signalResult.Conditions.Long.MACDCondition, "MACD"),
					formatConditionWithSymbol(signalResult.Conditions.Long.ParabolicSARCondition, "SAR"),
					signalResult.Conditions.Long.EMA200Value,
					signalResult.Conditions.Long.EMA200Diff,
					signalResult.Conditions.Long.MACDMACDLine,
					signalResult.Conditions.Long.MACDSignalLine,
					signalResult.Conditions.Long.MACDHistogram,
					signalResult.Conditions.Long.ParabolicSARValue,
					signalResult.Conditions.Long.ParabolicSARDiff),
				Inline: true,
			},
			{
				Name: "📉 SHORT",
				Value: fmt.Sprintf("```diff\n%s\n%s\n%s```\n```[EMA200]: %.2f (차이: %.2f)\n[MACD Line]: %.2f\n[Signal Line]: %.2f\n[Histogram]: %.2f\n[SAR]: %.2f (차이: %.2f)```",
					formatConditionWithSymbol(signalResult.Conditions.Short.EMA200Condition, "EMA200"),
					formatConditionWithSymbol(signalResult.Conditions.Short.MACDCondition, "MACD"),
					formatConditionWithSymbol(signalResult.Conditions.Short.ParabolicSARCondition, "SAR"),
					signalResult.Conditions.Short.EMA200Value,
					signalResult.Conditions.Short.EMA200Diff,
					signalResult.Conditions.Short.MACDMACDLine,
					signalResult.Conditions.Short.MACDSignalLine,
					signalResult.Conditions.Short.MACDHistogram,
					signalResult.Conditions.Short.ParabolicSARValue,
					signalResult.Conditions.Short.ParabolicSARDiff),
				Inline: true,
			},
		},
		Color:     notification.GetColorForDiscord(signalResult.Signal),
		Footer:    &notification.EmbedFooter{Text: "🤖 Assist Trading Bot"},
		Timestamp: timestamp.Format(time.RFC3339),
	}
}

func formatConditionWithSymbol(condition bool, text string) string {
	if condition {
		return fmt.Sprintf("✅ %s", text)
	}
	return fmt.Sprintf("❌ %s", text)
}

func processSignal(signalResult lib.SignalResult) error {
	log.Printf("Processing signal: %+v", signalResult)

	discordEmbed := generateDiscordEmbed(signalResult)
	if err := notification.SendDiscordAlert(discordEmbed, discordWebhookURL); err != nil {
		log.Printf("Error sending Discord alert: %v", err)
		return err
	}

	log.Println("Notifications sent successfully")
	return nil
}

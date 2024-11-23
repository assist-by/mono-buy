package main

import (
	"log"

	"github.com/assist-by/abmodule/notification"
	"github.com/assist-by/mono-buy/lib"
)

func processSignal(signalResult lib.SignalResult) error {
	// send notification
	sendNotification(signalResult)

	// send buy api

	return nil
}

// send notification
func sendNotification(signalResult lib.SignalResult) error {
	log.Printf("Processing signal: %+v", signalResult)

	// discord embedding
	discordEmbed := generateDiscordEmbed(signalResult)
	// discord로 알림 보내기
	if err := notification.SendDiscordAlert(discordEmbed, discordWebhookURL); err != nil {
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

package discord

import (
	"fmt"
	"strconv"
	"time"

	lib "github.com/assist-by/libStruct"
)

// SendTradeNotification sends a detailed trade notification to Discord
func (c *Client) SendTradeNotification(signalResult lib.SignalResult, orderSize float64, err error) error {
	var embed Embed

	// 주문 성공 시
	if err == nil {
		// 시그널 타입에 따른 색상과 이모지 설정
		var color int
		var emoji string
		switch signalResult.Signal {
		case lib.SIGNAL_LONG:
			color = ColorGreen
			emoji = "🚀"
		case lib.SIGNAL_SHORT:
			color = ColorRed
			emoji = "🔻"
		default:
			color = ColorBlue
			emoji = "⏺️"
		}

		// 수익률 계산
		var slPercent, tpPercent float64
		if signalResult.Signal == lib.SIGNAL_LONG {
			slPercent = (signalResult.StopLoss - signalResult.Price) / signalResult.Price * 100
			tpPercent = (signalResult.TakeProfit - signalResult.Price) / signalResult.Price * 100
		} else {
			slPercent = (signalResult.Price - signalResult.StopLoss) / signalResult.Price * 100
			tpPercent = (signalResult.Price - signalResult.TakeProfit) / signalResult.Price * 100
		}

		description := fmt.Sprintf(`**시간**: %s
**심볼**: %s/USDT
**포지션**: %s
**주문수량**: %.4f
**진입가**: $%.4f
**손절가**: $%.4f (%.2f%%)
**목표가**: $%.4f (%.2f%%)`,
			time.Unix(signalResult.Timestamp/1000, 0).Format("2006-01-02 15:04:05 KST"),
			signalResult.Symbol,
			strconv.Itoa(int(signalResult.Signal)),
			orderSize,
			signalResult.Price,
			signalResult.StopLoss,
			slPercent,
			signalResult.TakeProfit,
			tpPercent)

		embed = Embed{
			Title:       fmt.Sprintf("%s 주문 체결 완료", emoji),
			Description: description,
			Color:       color,
			Footer:      &EmbedFooter{Text: "🤖 Assist Trading Bot"},
			Timestamp:   time.Now().Format(time.RFC3339),
		}
	} else {
		// 주문 실패 시
		embed = Embed{
			Title: "⚠️ 주문 실패",
			Description: fmt.Sprintf(`**시간**: %s
**심볼**: %s/USDT
**에러**: %v`,
				time.Now().Format("2006-01-02 15:04:05 KST"),
				signalResult.Symbol,
				err),
			Color:     ColorRed,
			Footer:    &EmbedFooter{Text: "🤖 Assist Trading Bot"},
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	// Discord로 전송
	return c.Send(&embed)
}

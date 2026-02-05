package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// TelegramToken 全域 Bot Token
var TelegramToken string

// SetupWebhook 設定 Telegram Bot Webhook
// 原因：程式啟動時向 Telegram 註冊 webhook URL，讓訊息能推送到本服務
func SetupWebhook(token, webhookURL string) {
	TelegramToken = token

	url := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", token)
	body, _ := json.Marshal(map[string]string{
		"url": webhookURL,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("設定 Webhook 失敗: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Printf("Telegram Webhook 設定成功: %s", webhookURL)
	} else {
		log.Printf("Telegram Webhook 設定失敗，狀態碼: %d", resp.StatusCode)
	}
}

// SendMessage 透過 Telegram Bot API 發送訊息
// 原因：封裝發送邏輯，供 bot handler 呼叫
func SendMessage(chatID int64, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", TelegramToken)
	body, _ := json.Marshal(map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("發送訊息失敗: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

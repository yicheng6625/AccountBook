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

// InlineKeyboardButton Inline Keyboard 按鈕結構
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

// InlineKeyboardMarkup Inline Keyboard 整體結構
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

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

// SendMessage 發送純文字訊息
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

// SendMessageReturningID 發送純文字訊息並回傳訊息 ID
// 原因：發送提示訊息（如「請輸入金額：」）後需記錄 ID，使用者輸入後一併刪除
func SendMessageReturningID(chatID int64, text string) (int, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", TelegramToken)
	body, _ := json.Marshal(map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, fmt.Errorf("發送訊息失敗: %v", err)
	}
	defer resp.Body.Close()

	var result SendMessageResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Result.MessageID, nil
}

// SendMessageResponse sendMessage 的回應結構
// 原因：需要取得送出的訊息 ID，後續用於 editMessageText
type SendMessageResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
	} `json:"result"`
}

// SendMessageWithKeyboard 發送帶有 Inline Keyboard 的訊息，並回傳訊息 ID
// 原因：互動式新增流程的核心，顯示預覽資訊搭配可點擊的按鈕
func SendMessageWithKeyboard(chatID int64, text string, keyboard InlineKeyboardMarkup) (int, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", TelegramToken)
	body, _ := json.Marshal(map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"reply_markup": keyboard,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, fmt.Errorf("發送訊息失敗: %v", err)
	}
	defer resp.Body.Close()

	var result SendMessageResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Result.MessageID, nil
}

// EditMessageWithKeyboard 編輯已發送的訊息（更新文字與鍵盤）
// 原因：使用者修改欄位後，更新同一則預覽訊息而非發送新訊息，保持聊天室整潔
func EditMessageWithKeyboard(chatID int64, messageID int, text string, keyboard InlineKeyboardMarkup) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", TelegramToken)
	body, _ := json.Marshal(map[string]interface{}{
		"chat_id":      chatID,
		"message_id":   messageID,
		"text":         text,
		"reply_markup": keyboard,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("編輯訊息失敗: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

// EditMessageText 編輯已發送的訊息（僅更新文字，移除鍵盤）
// 原因：確認送出後，將預覽訊息替換為最終結果
func EditMessageText(chatID int64, messageID int, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", TelegramToken)
	body, _ := json.Marshal(map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("編輯訊息失敗: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

// AnswerCallbackQuery 回應 callback query（消除按鈕的 loading 狀態）
// 原因：Telegram 要求在收到 callback_query 後回應，否則按鈕會持續轉圈
func AnswerCallbackQuery(callbackQueryID string, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", TelegramToken)
	body, _ := json.Marshal(map[string]interface{}{
		"callback_query_id": callbackQueryID,
		"text":              text,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("回應 callback 失敗: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

// DeleteMessage 刪除訊息
// 原因：使用者輸入欄位值後，刪除使用者的輸入訊息保持整潔
func DeleteMessage(chatID int64, messageID int) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/deleteMessage", TelegramToken)
	body, _ := json.Marshal(map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("刪除訊息失敗: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

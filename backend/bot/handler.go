package bot

import (
	"accountbook/initializers"
	"accountbook/services"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TelegramUpdate Telegram 推播的更新結構
type TelegramUpdate struct {
	Message *TelegramMessage `json:"message"`
}

// TelegramMessage Telegram 訊息結構
type TelegramMessage struct {
	Chat *TelegramChat `json:"chat"`
	Text string        `json:"text"`
}

// TelegramChat Telegram 聊天室結構
type TelegramChat struct {
	ID int64 `json:"id"`
}

// HandleWebhook 處理 Telegram Webhook 推播
// 原因：接收 Telegram 訊息，依據內容分發到對應處理邏輯
func HandleWebhook(c *gin.Context) {
	var update TelegramUpdate
	if err := json.NewDecoder(c.Request.Body).Decode(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無法解析請求"})
		return
	}

	// 忽略非文字訊息
	if update.Message == nil || update.Message.Text == "" {
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	// 訊息路由分發
	switch {
	case text == "/新增格式" || text == "/start":
		services.SendMessage(chatID, FormatUsage())

	case text == "/查詢分類":
		services.SendMessage(chatID, FormatCategories())

	case text == "/查詢帳戶":
		services.SendMessage(chatID, FormatAccounts())

	case strings.HasPrefix(text, "/"):
		services.SendMessage(chatID, "未知指令，可用指令：\n/新增格式\n/查詢分類\n/查詢帳戶")

	default:
		// 嘗試解析為新增紀錄
		handleNewRecord(chatID, text)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// handleNewRecord 處理新增紀錄的訊息
// 原因：解析多行格式，新增紀錄並回覆結果
func handleNewRecord(chatID int64, text string) {
	record, err := ParseRecord(text)
	if err != nil {
		log.Printf("解析紀錄失敗: %v", err)
		services.SendMessage(chatID, FormatError())
		return
	}

	// 使用 Transaction 新增紀錄並更新帳戶餘額
	tx, err := initializers.DB.Begin()
	if err != nil {
		services.SendMessage(chatID, "系統錯誤，請稍後再試")
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	_, err = tx.Exec(
		"INSERT INTO records (date, account_id, type, amount, item, category_id, note, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		record.Date, record.AccountID, record.Type, record.Amount, record.Item, record.CategoryID, record.Note, now, now,
	)
	if err != nil {
		tx.Rollback()
		services.SendMessage(chatID, FormatError())
		return
	}

	// 更新帳戶餘額
	if record.Type == "支出" {
		tx.Exec("UPDATE accounts SET balance = balance - ?, updated_at = ? WHERE id = ?", record.Amount, now, record.AccountID)
	} else {
		tx.Exec("UPDATE accounts SET balance = balance + ?, updated_at = ? WHERE id = ?", record.Amount, now, record.AccountID)
	}

	if err = tx.Commit(); err != nil {
		services.SendMessage(chatID, "系統錯誤，請稍後再試")
		return
	}

	// 查詢帳戶與分類名稱
	var accountName, categoryName string
	initializers.DB.QueryRow("SELECT name FROM accounts WHERE id = ?", record.AccountID).Scan(&accountName)
	initializers.DB.QueryRow("SELECT name FROM categories WHERE id = ?", record.CategoryID).Scan(&categoryName)

	// 回覆成功訊息
	reply := FormatSuccess(record.Date, accountName, record.Type, record.Amount, record.Item, categoryName, record.Note)
	services.SendMessage(chatID, reply)
}

package bot

import (
	"accountbook/initializers"
	"accountbook/services"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// === Telegram 資料結構 ===

// TelegramUpdate Telegram Webhook 推播的完整結構
// 原因：需同時處理一般訊息 (message) 和按鈕回調 (callback_query)
type TelegramUpdate struct {
	Message       *TelegramMessage       `json:"message"`
	CallbackQuery *TelegramCallbackQuery `json:"callback_query"`
}

type TelegramMessage struct {
	MessageID int           `json:"message_id"`
	Chat      *TelegramChat `json:"chat"`
	Text      string        `json:"text"`
}

type TelegramChat struct {
	ID int64 `json:"id"`
}

type TelegramCallbackQuery struct {
	ID      string           `json:"id"`
	Message *TelegramMessage `json:"message"`
	Data    string           `json:"data"`
}

// === Webhook 入口 ===

// HandleWebhook 處理 Telegram Webhook 推播
func HandleWebhook(c *gin.Context) {
	var update TelegramUpdate
	if err := json.NewDecoder(c.Request.Body).Decode(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無法解析請求"})
		return
	}

	// 處理按鈕回調（Callback Query）
	if update.CallbackQuery != nil {
		handleCallbackQuery(update.CallbackQuery)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	// 處理一般訊息
	if update.Message != nil && update.Message.Text != "" {
		handleMessage(update.Message)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// === 一般訊息處理 ===

func handleMessage(msg *TelegramMessage) {
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)

	// 指令處理
	switch {
	case text == "/start":
		services.SendMessage(chatID, FormatUsage())
		return

	case text == "/查詢分類" || text == "/categories" || text == "/分類":
		services.SendMessage(chatID, FormatCategories())
		return

	case text == "/查詢帳戶" || text == "/accounts" || text == "/帳戶":
		services.SendMessage(chatID, FormatAccounts())
		return

	case text == "/new" || text == "/記帳":
		startNewRecord(chatID)
		return

	case text == "/recent" || text == "/最近":
		handleRecentRecords(chatID, 0)
		return

	case text == "/transfer" || text == "/轉帳":
		startTransfer(chatID)
		return

	case text == "/cancel" || text == "/取消":
		DeleteSession(chatID)
		services.SendMessage(chatID, "已取消")
		return

	case strings.HasPrefix(text, "/"):
		services.SendMessage(chatID, "未知指令，可用指令：/start、/new、/transfer、/recent、/查詢帳戶、/查詢分類")
		return
	}

	// 檢查是否有進行中的會話（等待使用者輸入欄位值）
	session := GetSession(chatID)
	if session != nil && session.State != StatePreview && session.State != StateIdle && session.State != StateTransferPreview {
		handleFieldInput(chatID, msg.MessageID, session, text)
		return
	}

	// 非指令、無會話 → 嘗試解析快捷輸入後開始新增紀錄流程
	startNewRecordWithQuickInput(chatID, text)
}

// startNewRecord 啟動互動式新增紀錄流程
// 原因：建立帶有預設值的會話，發送預覽訊息搭配 Inline Keyboard
func startNewRecord(chatID int64) {
	session := NewSession(chatID)

	text := FormatPreview(session)
	keyboard := BuildPreviewKeyboard(session)

	msgID, err := services.SendMessageWithKeyboard(chatID, text, keyboard)
	if err != nil {
		log.Printf("發送預覽訊息失敗: %v", err)
		return
	}

	session.MessageID = msgID
}

// handleFieldInput 處理使用者輸入的欄位值
// 原因：使用者點擊「修改日期」等按鈕後，Bot 等待使用者輸入文字
func handleFieldInput(chatID int64, userMsgID int, session *Session, text string) {
	switch session.State {
	case StateEditDate:
		date, err := parseDate(text)
		if err != nil {
			services.SendMessage(chatID, "❌ 日期格式錯誤，請輸入如：今天、昨天、2026/01/15、01/15")
			return
		}
		session.Date = date

	case StateEditAmt:
		amount, err := parseAmount(text)
		if err != nil {
			services.SendMessage(chatID, "❌ 請輸入正確的金額（數字）")
			return
		}
		session.Amount = amount

	case StateEditItem:
		session.Item = text

	case StateEditNote:
		session.Note = text

	// 轉帳專用狀態
	case StateTransferAmt:
		amount, err := parseAmount(text)
		if err != nil {
			services.SendMessage(chatID, "❌ 請輸入正確的金額（數字）")
			return
		}
		session.Amount = amount

	case StateTransferNote:
		session.Note = text
	}

	// 刪除「請輸入XXX：」提示訊息
	if session.PromptMsgID > 0 {
		services.DeleteMessage(chatID, session.PromptMsgID)
		session.PromptMsgID = 0
	}

	// 刪除使用者的輸入訊息，保持聊天室整潔
	services.DeleteMessage(chatID, userMsgID)

	// 根據模式回到對應的預覽狀態
	if session.Mode == ModeTransfer {
		session.State = StateTransferPreview
		updateTransferPreview(chatID, session)
	} else {
		session.State = StatePreview
		updatePreview(chatID, session)
	}
}

// updatePreview 更新預覽訊息的文字與鍵盤
func updatePreview(chatID int64, session *Session) {
	text := FormatPreview(session)
	keyboard := BuildPreviewKeyboard(session)

	if session.MessageID > 0 {
		services.EditMessageWithKeyboard(chatID, session.MessageID, text, keyboard)
	}
}

// === Callback Query 處理（按鈕點擊）===

func handleCallbackQuery(cq *TelegramCallbackQuery) {
	chatID := cq.Message.Chat.ID
	data := cq.Data

	// 先回應 callback（消除按鈕 loading）
	services.AnswerCallbackQuery(cq.ID, "")

	// 處理翻頁按鈕（不需要會話）
	if strings.HasPrefix(data, "recent_page_") {
		offsetStr := strings.TrimPrefix(data, "recent_page_")
		offset, _ := strconv.Atoi(offsetStr)
		const pageSize = 5
		text, total := FormatRecentRecords(offset, pageSize)
		keyboard := BuildPaginationKeyboard(offset, pageSize, total)
		services.EditMessageWithKeyboard(chatID, cq.Message.MessageID, text, keyboard)
		return
	}

	session := GetSession(chatID)

	// 若無會話但收到 callback，可能是過期的按鈕
	if session == nil {
		if data == "new_record" {
			startNewRecord(chatID)
			return
		}
		services.AnswerCallbackQuery(cq.ID, "此操作已過期，請重新開始")
		return
	}

	// 轉帳模式的 callback 處理
	if session.Mode == ModeTransfer {
		handleTransferCallback(chatID, cq, session, data)
		return
	}

	switch {
	// 編輯日期：顯示快捷日期選擇鍵盤
	case data == "edit_date":
		keyboard := BuildDateKeyboard()
		services.EditMessageWithKeyboard(chatID, session.MessageID,
			"📅 選擇日期，或直接輸入（如：2026/01/15、01/15）", keyboard)
		session.State = StateEditDate

	// 快捷日期選擇
	case strings.HasPrefix(data, "set_date_"):
		offsetStr := strings.TrimPrefix(data, "set_date_")
		offset, _ := strconv.Atoi(offsetStr)
		session.Date = time.Now().AddDate(0, 0, offset).Format("2006-01-02")
		session.State = StatePreview
		updatePreview(chatID, session)

	// 編輯帳戶：顯示帳戶選擇鍵盤
	case data == "edit_account":
		keyboard := BuildAccountKeyboard()
		services.EditMessageWithKeyboard(chatID, session.MessageID,
			"🏦 選擇帳戶：", keyboard)

	// 選擇帳戶
	case strings.HasPrefix(data, "set_account_"):
		idStr := strings.TrimPrefix(data, "set_account_")
		id, _ := strconv.Atoi(idStr)
		session.AccountID = id
		session.State = StatePreview
		updatePreview(chatID, session)

	// 編輯類型：顯示收入/支出選擇
	case data == "edit_type":
		keyboard := BuildTypeKeyboard()
		services.EditMessageWithKeyboard(chatID, session.MessageID,
			"💱 選擇類型：", keyboard)

	// 選擇類型
	case strings.HasPrefix(data, "set_type_"):
		session.Type = strings.TrimPrefix(data, "set_type_")
		session.State = StatePreview
		updatePreview(chatID, session)

	// 編輯金額：發送提示訊息，等待使用者輸入
	case data == "edit_amount":
		session.State = StateEditAmt
		promptID, _ := services.SendMessageReturningID(chatID, "💰 請輸入金額：")
		session.PromptMsgID = promptID

	// 編輯項目：發送提示訊息，等待使用者輸入
	case data == "edit_item":
		session.State = StateEditItem
		promptID, _ := services.SendMessageReturningID(chatID, "📝 請輸入項目名稱：")
		session.PromptMsgID = promptID

	// 編輯分類：顯示分類選擇鍵盤
	case data == "edit_category":
		keyboard := BuildCategoryKeyboard()
		services.EditMessageWithKeyboard(chatID, session.MessageID,
			"🏷 選擇分類：", keyboard)

	// 選擇分類
	case strings.HasPrefix(data, "set_category_"):
		idStr := strings.TrimPrefix(data, "set_category_")
		id, _ := strconv.Atoi(idStr)
		session.CategoryID = id
		session.State = StatePreview
		updatePreview(chatID, session)

	// 編輯備註：發送提示訊息，等待使用者輸入
	case data == "edit_note":
		session.State = StateEditNote
		promptID, _ := services.SendMessageReturningID(chatID, "📌 請輸入備註（輸入「無」可清除）：")
		session.PromptMsgID = promptID

	// 確認送出
	case data == "confirm":
		handleConfirm(chatID, session)

	// 取消
	case data == "cancel":
		DeleteSession(chatID)
		services.EditMessageText(chatID, session.MessageID, "❌ 已取消新增紀錄")
	}
}

// startNewRecordWithQuickInput 解析快捷輸入並帶入欄位後開始新增紀錄
// 支援格式：
//   - 純數字（如 "150"）→ 帶入金額
//   - "文字 數字"（如 "午餐 150"）→ 帶入項目名稱 + 金額
//   - "數字 文字"（如 "150 午餐"）→ 帶入金額 + 項目名稱
func startNewRecordWithQuickInput(chatID int64, text string) {
	session := NewSession(chatID)

	// 嘗試解析快捷格式
	item, amount := parseQuickInput(text)
	if amount > 0 {
		session.Amount = amount
	}
	if item != "" {
		session.Item = item
	}

	text2 := FormatPreview(session)
	keyboard := BuildPreviewKeyboard(session)

	msgID, err := services.SendMessageWithKeyboard(chatID, text2, keyboard)
	if err != nil {
		log.Printf("發送預覽訊息失敗: %v", err)
		return
	}

	session.MessageID = msgID
}

// parseQuickInput 解析快捷輸入文字，回傳項目名稱與金額
func parseQuickInput(text string) (item string, amount float64) {
	// 純數字 → 金額
	if a, err := strconv.ParseFloat(text, 64); err == nil && a > 0 {
		return "", a
	}

	// 以空白分割，嘗試「文字 數字」或「數字 文字」
	parts := strings.Fields(text)
	if len(parts) == 2 {
		// 「文字 數字」
		if a, err := strconv.ParseFloat(parts[1], 64); err == nil && a > 0 {
			return parts[0], a
		}
		// 「數字 文字」
		if a, err := strconv.ParseFloat(parts[0], 64); err == nil && a > 0 {
			return parts[1], a
		}
	}

	return "", 0
}

// handleRecentRecords 查詢最近紀錄並發送帶翻頁按鈕的訊息
func handleRecentRecords(chatID int64, offset int) {
	const pageSize = 5
	text, total := FormatRecentRecords(offset, pageSize)
	keyboard := BuildPaginationKeyboard(offset, pageSize, total)

	if len(keyboard.InlineKeyboard) > 0 {
		services.SendMessageWithKeyboard(chatID, text, keyboard)
	} else {
		services.SendMessage(chatID, text)
	}
}

// === 轉帳功能 ===

// startTransfer 啟動轉帳流程
func startTransfer(chatID int64) {
	session := NewTransferSession(chatID)

	text := FormatTransferPreview(session)
	keyboard := BuildTransferKeyboard(session)

	msgID, err := services.SendMessageWithKeyboard(chatID, text, keyboard)
	if err != nil {
		log.Printf("發送轉帳預覽失敗: %v", err)
		return
	}

	session.MessageID = msgID
}

// updateTransferPreview 更新轉帳預覽訊息
func updateTransferPreview(chatID int64, session *Session) {
	text := FormatTransferPreview(session)
	keyboard := BuildTransferKeyboard(session)

	if session.MessageID > 0 {
		services.EditMessageWithKeyboard(chatID, session.MessageID, text, keyboard)
	}
}

// handleTransferCallback 處理轉帳相關的 callback
func handleTransferCallback(chatID int64, cq *TelegramCallbackQuery, session *Session, data string) {
	switch {
	case data == "transfer_edit_from":
		keyboard := BuildTransferAccountKeyboard("transfer_from_")
		services.EditMessageWithKeyboard(chatID, session.MessageID,
			"🏦 選擇轉出帳戶：", keyboard)

	case strings.HasPrefix(data, "transfer_from_"):
		idStr := strings.TrimPrefix(data, "transfer_from_")
		id, _ := strconv.Atoi(idStr)
		session.AccountID = id
		session.State = StateTransferPreview
		updateTransferPreview(chatID, session)

	case data == "transfer_edit_to":
		keyboard := BuildTransferAccountKeyboard("transfer_to_")
		services.EditMessageWithKeyboard(chatID, session.MessageID,
			"🏦 選擇轉入帳戶：", keyboard)

	case strings.HasPrefix(data, "transfer_to_"):
		idStr := strings.TrimPrefix(data, "transfer_to_")
		id, _ := strconv.Atoi(idStr)
		session.ToAccountID = id
		session.State = StateTransferPreview
		updateTransferPreview(chatID, session)

	case data == "transfer_edit_amount":
		session.State = StateTransferAmt
		promptID, _ := services.SendMessageReturningID(chatID, "💰 請輸入轉帳金額：")
		session.PromptMsgID = promptID

	case data == "transfer_edit_note":
		session.State = StateTransferNote
		promptID, _ := services.SendMessageReturningID(chatID, "📌 請輸入備註（輸入「無」可清除）：")
		session.PromptMsgID = promptID

	case data == "transfer_confirm":
		handleTransferConfirm(chatID, session)

	case data == "cancel":
		DeleteSession(chatID)
		services.EditMessageText(chatID, session.MessageID, "❌ 已取消轉帳")
	}
}

// handleTransferConfirm 確認轉帳
func handleTransferConfirm(chatID int64, session *Session) {
	if session.Amount <= 0 {
		updateTransferPreview(chatID, session)
		services.SendMessage(chatID, "⚠️ 請先填寫轉帳金額")
		return
	}

	if session.AccountID == session.ToAccountID {
		updateTransferPreview(chatID, session)
		services.SendMessage(chatID, "⚠️ 轉出與轉入帳戶不能相同")
		return
	}

	if session.Note == "無" {
		session.Note = ""
	}

	tx, err := initializers.DB.Begin()
	if err != nil {
		services.SendMessage(chatID, "系統錯誤，請稍後再試")
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	// 扣除轉出帳戶餘額
	_, err = tx.Exec("UPDATE accounts SET balance = balance - ?, updated_at = ? WHERE id = ?",
		session.Amount, now, session.AccountID)
	if err != nil {
		tx.Rollback()
		services.SendMessage(chatID, "轉帳失敗")
		return
	}

	// 增加轉入帳戶餘額
	_, err = tx.Exec("UPDATE accounts SET balance = balance + ?, updated_at = ? WHERE id = ?",
		session.Amount, now, session.ToAccountID)
	if err != nil {
		tx.Rollback()
		services.SendMessage(chatID, "轉帳失敗")
		return
	}

	if err = tx.Commit(); err != nil {
		services.SendMessage(chatID, "系統錯誤，請稍後再試")
		return
	}

	fromName := resolveAccountName(session.AccountID)
	toName := resolveAccountName(session.ToAccountID)

	successMsg := FormatTransferSuccess(fromName, toName, session.Amount, session.Note)
	services.EditMessageText(chatID, session.MessageID, successMsg)

	DeleteSession(chatID)
}

// handleConfirm 確認送出紀錄
// 原因：驗證必填欄位後，寫入資料庫並更新帳戶餘額
func handleConfirm(chatID int64, session *Session) {
	// 驗證必填欄位
	if session.Amount <= 0 {
		services.AnswerCallbackQuery("", "請先填寫金額")
		updatePreview(chatID, session)
		services.SendMessage(chatID, "⚠️ 請先點擊「💰 金額」填寫金額")
		return
	}
	if session.Item == "" {
		updatePreview(chatID, session)
		services.SendMessage(chatID, "⚠️ 請先點擊「📝 項目」填寫項目名稱")
		return
	}

	// 備註為「無」時清除
	if session.Note == "無" {
		session.Note = ""
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
		session.Date, session.AccountID, session.Type, session.Amount, session.Item, session.CategoryID, session.Note, now, now,
	)
	if err != nil {
		tx.Rollback()
		services.SendMessage(chatID, "新增紀錄失敗")
		log.Printf("新增紀錄失敗: %v", err)
		return
	}

	// 更新帳戶餘額
	if session.Type == "支出" {
		tx.Exec("UPDATE accounts SET balance = balance - ?, updated_at = ? WHERE id = ?", session.Amount, now, session.AccountID)
	} else {
		tx.Exec("UPDATE accounts SET balance = balance + ?, updated_at = ? WHERE id = ?", session.Amount, now, session.AccountID)
	}

	if err = tx.Commit(); err != nil {
		services.SendMessage(chatID, "系統錯誤，請稍後再試")
		return
	}

	// 取得名稱用於回覆
	accountName := resolveAccountName(session.AccountID)
	categoryName := resolveCategoryName(session.CategoryID)

	// 更新預覽訊息為成功訊息（移除鍵盤）
	successMsg := FormatSuccess(session.Date, accountName, session.Type, session.Amount, session.Item, categoryName, session.Note)
	services.EditMessageText(chatID, session.MessageID, successMsg)

	// 清除會話
	DeleteSession(chatID)
}


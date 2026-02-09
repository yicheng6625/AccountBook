package bot

import (
	"accountbook/initializers"
	"sync"
	"time"
)

// SessionState 會話狀態列舉
// 原因：追蹤使用者在新增紀錄流程中的目前步驟
type SessionState string

const (
	StateIdle     SessionState = ""         // 無進行中的操作
	StatePreview  SessionState = "preview"  // 預覽中，等待使用者點擊欄位修改或確認送出
	StateEditDate SessionState = "date"     // 等待輸入日期
	StateEditAmt  SessionState = "amount"   // 等待輸入金額
	StateEditItem SessionState = "item"     // 等待輸入項目名稱
	StateEditNote SessionState = "note"     // 等待輸入備註

	// 轉帳專用狀態
	StateTransferPreview SessionState = "transfer_preview" // 轉帳預覽
	StateTransferAmt     SessionState = "transfer_amount"  // 等待輸入轉帳金額
	StateTransferNote    SessionState = "transfer_note"    // 等待輸入轉帳備註
)

// SessionMode 會話模式
type SessionMode string

const (
	ModeRecord   SessionMode = "record"   // 記帳模式
	ModeTransfer SessionMode = "transfer" // 轉帳模式
)

// Session 單一使用者的會話狀態
type Session struct {
	Mode        SessionMode
	State       SessionState
	Date        string  // 日期
	AccountID   int     // 帳戶 ID（記帳用，或轉帳來源）
	Type        string  // 收入/支出
	Amount      float64 // 金額
	Item        string  // 項目名稱
	CategoryID  int     // 分類 ID
	Note        string  // 備註
	MessageID   int     // 上一則預覽訊息的 ID（用於編輯訊息）
	PromptMsgID int     // 「請輸入XXX：」提示訊息的 ID（原因：使用者輸入後需一併刪除）
	ToAccountID int     // 轉帳目標帳戶 ID
	UpdatedAt   time.Time
}

// sessionStore 全域會話儲存（以 chatID 為 key）
// 原因：Telegram Bot 是無狀態的 Webhook，需自行管理使用者的操作狀態
var (
	sessionStore = make(map[int64]*Session)
	sessionMu    sync.RWMutex
)

// GetSession 取得使用者的會話，若不存在則建立新的
func GetSession(chatID int64) *Session {
	sessionMu.RLock()
	s, ok := sessionStore[chatID]
	sessionMu.RUnlock()
	if ok {
		return s
	}
	return nil
}

// NewSession 建立新會話並帶入預設值
// 原因：開始新增紀錄時，預先填入今天日期、預設帳戶、支出等預設值
func NewSession(chatID int64) *Session {
	s := &Session{
		Mode:       ModeRecord,
		State:      StatePreview,
		Date:       time.Now().Format("2006-01-02"),
		AccountID:  getDefaultAccountID(),
		Type:       "支出",
		Amount:     0,
		Item:       "",
		CategoryID: getDefaultCategoryID(),
		Note:       "",
		UpdatedAt:  time.Now(),
	}
	sessionMu.Lock()
	sessionStore[chatID] = s
	sessionMu.Unlock()
	return s
}

// NewTransferSession 建立轉帳會話
func NewTransferSession(chatID int64) *Session {
	s := &Session{
		Mode:      ModeTransfer,
		State:     StateTransferPreview,
		Date:      time.Now().Format("2006-01-02"),
		AccountID: getDefaultAccountID(),
		Amount:    0,
		Note:      "",
		UpdatedAt: time.Now(),
	}
	// 取得第二個帳戶作為預設轉入帳戶
	s.ToAccountID = getSecondAccountID(s.AccountID)
	sessionMu.Lock()
	sessionStore[chatID] = s
	sessionMu.Unlock()
	return s
}

// getSecondAccountID 取得非指定帳戶的第一個帳戶 ID
func getSecondAccountID(excludeID int) int {
	var id int
	err := initializers.DB.QueryRow("SELECT id FROM accounts WHERE id != ? ORDER BY sort_order LIMIT 1", excludeID).Scan(&id)
	if err != nil {
		return excludeID
	}
	return id
}

// DeleteSession 清除使用者的會話
func DeleteSession(chatID int64) {
	sessionMu.Lock()
	delete(sessionStore, chatID)
	sessionMu.Unlock()
}

// getDefaultCategoryID 取得第一個分類的 ID 作為預設值
func getDefaultCategoryID() int {
	var id int
	err := initializers.DB.QueryRow("SELECT id FROM categories ORDER BY sort_order LIMIT 1").Scan(&id)
	if err != nil {
		return 1
	}
	return id
}

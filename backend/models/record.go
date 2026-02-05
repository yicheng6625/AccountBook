package models

// Record 記帳紀錄模型
// 原因：對應 records 資料表，為系統核心資料結構
type Record struct {
	ID         int     `json:"id"`
	Date       string  `json:"date"`
	AccountID  int     `json:"account_id"`
	Type       string  `json:"type"`
	Amount     float64 `json:"amount"`
	Item       string  `json:"item"`
	CategoryID int     `json:"category_id"`
	Note       string  `json:"note"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// RecordWithNames 帶有帳戶與分類名稱的紀錄
// 原因：API 回應時需要顯示名稱而非僅 ID
type RecordWithNames struct {
	ID           int     `json:"id"`
	Date         string  `json:"date"`
	AccountID    int     `json:"account_id"`
	AccountName  string  `json:"account_name"`
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	Item         string  `json:"item"`
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Note         string  `json:"note"`
}

// RecordInput 新增/更新紀錄的輸入資料
// 原因：分離輸入與輸出結構，避免欄位混淆
type RecordInput struct {
	Date       string  `json:"date" binding:"required"`
	AccountID  int     `json:"account_id" binding:"required"`
	Type       string  `json:"type"`
	Amount     float64 `json:"amount" binding:"required"`
	Item       string  `json:"item" binding:"required"`
	CategoryID int     `json:"category_id" binding:"required"`
	Note       string  `json:"note"`
}

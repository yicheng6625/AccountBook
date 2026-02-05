package models

// Account 帳戶模型
// 原因：對應 accounts 資料表，記錄不同支付方式及各自餘額
type Account struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Balance   float64 `json:"balance"`
	SortOrder int     `json:"sort_order"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

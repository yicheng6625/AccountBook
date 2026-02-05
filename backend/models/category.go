package models

// Category 分類模型
// 原因：對應 categories 資料表，提供自訂分類功能
type Category struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

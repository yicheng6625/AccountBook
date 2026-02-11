package controllers

import (
	"accountbook/initializers"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TransferInput 轉帳輸入資料
type TransferInput struct {
	Date          string  `json:"date"`
	FromAccountID int     `json:"from_account_id" binding:"required"`
	ToAccountID   int     `json:"to_account_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	Note          string  `json:"note"`
}

// CreateTransfer 新增轉帳紀錄
// 更新兩個帳戶餘額，並建立兩筆 records 以便在日曆中顯示
func CreateTransfer(c *gin.Context) {
	var input TransferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "輸入格式錯誤，請確認必填欄位"})
		return
	}

	if input.FromAccountID == input.ToAccountID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "轉出與轉入帳戶不能相同"})
		return
	}

	if input.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "金額必須大於 0"})
		return
	}

	// 預設日期為今天
	if input.Date == "" {
		input.Date = time.Now().Format("2006-01-02")
	}

	// 取得帳戶名稱
	var fromName, toName string
	initializers.DB.QueryRow("SELECT name FROM accounts WHERE id = ?", input.FromAccountID).Scan(&fromName)
	initializers.DB.QueryRow("SELECT name FROM accounts WHERE id = ?", input.ToAccountID).Scan(&toName)

	if fromName == "" || toName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帳戶不存在"})
		return
	}

	// 取得或建立「轉帳」分類
	categoryID := getTransferCategoryID()

	tx, err := initializers.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "交易開始失敗"})
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	// 扣除轉出帳戶餘額
	_, err = tx.Exec("UPDATE accounts SET balance = balance - ?, updated_at = ? WHERE id = ?",
		input.Amount, now, input.FromAccountID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "轉帳失敗"})
		return
	}

	// 增加轉入帳戶餘額
	_, err = tx.Exec("UPDATE accounts SET balance = balance + ?, updated_at = ? WHERE id = ?",
		input.Amount, now, input.ToAccountID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "轉帳失敗"})
		return
	}

	// 建立轉出紀錄（支出）
	outItem := fmt.Sprintf("轉帳至 %s", toName)
	_, err = tx.Exec(
		"INSERT INTO records (date, account_id, type, amount, item, category_id, note, created_at, updated_at) VALUES (?, ?, '支出', ?, ?, ?, ?, ?, ?)",
		input.Date, input.FromAccountID, input.Amount, outItem, categoryID, input.Note, now, now,
	)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "建立轉帳紀錄失敗"})
		return
	}

	// 建立轉入紀錄（收入）
	inItem := fmt.Sprintf("從 %s 轉入", fromName)
	_, err = tx.Exec(
		"INSERT INTO records (date, account_id, type, amount, item, category_id, note, created_at, updated_at) VALUES (?, ?, '收入', ?, ?, ?, ?, ?, ?)",
		input.Date, input.ToAccountID, input.Amount, inItem, categoryID, input.Note, now, now,
	)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "建立轉帳紀錄失敗"})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "交易提交失敗"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "轉帳成功",
		"date":         input.Date,
		"from_account": fromName,
		"to_account":   toName,
		"amount":       input.Amount,
		"note":         input.Note,
	})
}

// getTransferCategoryID 取得或建立「轉帳」分類的 ID
func getTransferCategoryID() int {
	var id int
	err := initializers.DB.QueryRow("SELECT id FROM categories WHERE name = '轉帳'").Scan(&id)
	if err == nil {
		return id
	}

	// 不存在則建立
	result, err := initializers.DB.Exec("INSERT INTO categories (name, sort_order) VALUES ('轉帳', 999)")
	if err != nil {
		return 1 // fallback
	}
	newID, _ := result.LastInsertId()
	return int(newID)
}

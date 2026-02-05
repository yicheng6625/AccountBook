package controllers

import (
	"accountbook/initializers"
	"accountbook/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetRecords 查詢紀錄
// 原因：支援兩種查詢模式 - 按日期（首頁紀錄列表）或按月份（行事曆每日摘要）
func GetRecords(c *gin.Context) {
	date := c.Query("date")
	month := c.Query("month")

	if date != "" {
		getRecordsByDate(c, date)
	} else if month != "" {
		getRecordsByMonth(c, month)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請提供 date 或 month 參數"})
	}
}

// getRecordsByDate 查詢指定日期的紀錄列表
// 原因：首頁點擊行事曆日期時，顯示當日所有紀錄
func getRecordsByDate(c *gin.Context, date string) {
	rows, err := initializers.DB.Query(`
		SELECT r.id, r.date, r.account_id, a.name, r.type, r.amount, r.item, r.category_id, c.name, r.note
		FROM records r
		JOIN accounts a ON r.account_id = a.id
		JOIN categories c ON r.category_id = c.id
		WHERE r.date = ?
		ORDER BY r.created_at DESC
	`, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢紀錄失敗"})
		return
	}
	defer rows.Close()

	var records []models.RecordWithNames
	var totalIncome, totalExpense float64

	for rows.Next() {
		var r models.RecordWithNames
		if err := rows.Scan(&r.ID, &r.Date, &r.AccountID, &r.AccountName, &r.Type, &r.Amount, &r.Item, &r.CategoryID, &r.CategoryName, &r.Note); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "讀取紀錄資料失敗"})
			return
		}
		if r.Type == "收入" {
			totalIncome += r.Amount
		} else {
			totalExpense += r.Amount
		}
		records = append(records, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"date":          date,
		"records":       records,
		"total_income":  totalIncome,
		"total_expense": totalExpense,
	})
}

// getRecordsByMonth 查詢指定月份的每日摘要
// 原因：行事曆需要知道哪些日期有紀錄，以及每日收支金額
func getRecordsByMonth(c *gin.Context, month string) {
	rows, err := initializers.DB.Query(`
		SELECT date, type, SUM(amount) as total
		FROM records
		WHERE strftime('%Y-%m', date) = ?
		GROUP BY date, type
		ORDER BY date
	`, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢月份資料失敗"})
		return
	}
	defer rows.Close()

	// 每日摘要：日期 -> {income, expense}
	type DailySummary struct {
		Income  float64 `json:"income"`
		Expense float64 `json:"expense"`
	}
	dailySummary := make(map[string]*DailySummary)

	for rows.Next() {
		var date, recordType string
		var total float64
		if err := rows.Scan(&date, &recordType, &total); err != nil {
			continue
		}
		if _, ok := dailySummary[date]; !ok {
			dailySummary[date] = &DailySummary{}
		}
		if recordType == "收入" {
			dailySummary[date].Income = total
		} else {
			dailySummary[date].Expense = total
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"month":         month,
		"daily_summary": dailySummary,
	})
}

// GetRecord 取得單筆紀錄詳情
// 原因：進入編輯頁時需要取得完整紀錄資料
func GetRecord(c *gin.Context) {
	id := c.Param("id")

	var r models.RecordWithNames
	err := initializers.DB.QueryRow(`
		SELECT r.id, r.date, r.account_id, a.name, r.type, r.amount, r.item, r.category_id, c.name, r.note
		FROM records r
		JOIN accounts a ON r.account_id = a.id
		JOIN categories c ON r.category_id = c.id
		WHERE r.id = ?
	`, id).Scan(&r.ID, &r.Date, &r.AccountID, &r.AccountName, &r.Type, &r.Amount, &r.Item, &r.CategoryID, &r.CategoryName, &r.Note)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到該紀錄"})
		return
	}

	c.JSON(http.StatusOK, r)
}

// CreateRecord 新增紀錄
// 原因：新增紀錄時需同步更新帳戶餘額，使用 Transaction 確保資料一致性
func CreateRecord(c *gin.Context) {
	var input models.RecordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "輸入格式錯誤，請確認必填欄位"})
		return
	}

	// 預設類型為支出
	if input.Type == "" {
		input.Type = "支出"
	}

	// 使用 Transaction 確保紀錄新增與帳戶餘額同步
	tx, err := initializers.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "交易開始失敗"})
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	// 新增紀錄
	result, err := tx.Exec(
		"INSERT INTO records (date, account_id, type, amount, item, category_id, note, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		input.Date, input.AccountID, input.Type, input.Amount, input.Item, input.CategoryID, input.Note, now, now,
	)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "新增紀錄失敗"})
		return
	}

	// 更新帳戶餘額：支出扣款、收入加款
	if input.Type == "支出" {
		_, err = tx.Exec("UPDATE accounts SET balance = balance - ?, updated_at = ? WHERE id = ?", input.Amount, now, input.AccountID)
	} else {
		_, err = tx.Exec("UPDATE accounts SET balance = balance + ?, updated_at = ? WHERE id = ?", input.Amount, now, input.AccountID)
	}
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新帳戶餘額失敗"})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "交易提交失敗"})
		return
	}

	recordID, _ := result.LastInsertId()

	// 查詢帳戶與分類名稱回傳
	var accountName, categoryName string
	initializers.DB.QueryRow("SELECT name FROM accounts WHERE id = ?", input.AccountID).Scan(&accountName)
	initializers.DB.QueryRow("SELECT name FROM categories WHERE id = ?", input.CategoryID).Scan(&categoryName)

	c.JSON(http.StatusCreated, gin.H{
		"id":            recordID,
		"date":          input.Date,
		"account_name":  accountName,
		"type":          input.Type,
		"amount":        input.Amount,
		"item":          input.Item,
		"category_name": categoryName,
		"note":          input.Note,
	})
}

// UpdateRecord 更新紀錄
// 原因：需回滾舊紀錄對帳戶餘額的影響，再套用新值
func UpdateRecord(c *gin.Context) {
	id := c.Param("id")

	var input models.RecordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "輸入格式錯誤"})
		return
	}

	if input.Type == "" {
		input.Type = "支出"
	}

	// 先查詢舊紀錄，用於回滾帳戶餘額
	var oldAccountID int
	var oldType string
	var oldAmount float64
	err := initializers.DB.QueryRow("SELECT account_id, type, amount FROM records WHERE id = ?", id).
		Scan(&oldAccountID, &oldType, &oldAmount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到該紀錄"})
		return
	}

	tx, err := initializers.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "交易開始失敗"})
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	// 回滾舊帳戶餘額
	if oldType == "支出" {
		tx.Exec("UPDATE accounts SET balance = balance + ?, updated_at = ? WHERE id = ?", oldAmount, now, oldAccountID)
	} else {
		tx.Exec("UPDATE accounts SET balance = balance - ?, updated_at = ? WHERE id = ?", oldAmount, now, oldAccountID)
	}

	// 更新紀錄
	_, err = tx.Exec(
		"UPDATE records SET date=?, account_id=?, type=?, amount=?, item=?, category_id=?, note=?, updated_at=? WHERE id=?",
		input.Date, input.AccountID, input.Type, input.Amount, input.Item, input.CategoryID, input.Note, now, id,
	)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新紀錄失敗"})
		return
	}

	// 套用新帳戶餘額
	if input.Type == "支出" {
		tx.Exec("UPDATE accounts SET balance = balance - ?, updated_at = ? WHERE id = ?", input.Amount, now, input.AccountID)
	} else {
		tx.Exec("UPDATE accounts SET balance = balance + ?, updated_at = ? WHERE id = ?", input.Amount, now, input.AccountID)
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "交易提交失敗"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteRecord 刪除紀錄
// 原因：刪除紀錄時需回滾對帳戶餘額的影響
func DeleteRecord(c *gin.Context) {
	id := c.Param("id")

	// 先查詢紀錄以回滾餘額
	var accountID int
	var recordType string
	var amount float64
	err := initializers.DB.QueryRow("SELECT account_id, type, amount FROM records WHERE id = ?", id).
		Scan(&accountID, &recordType, &amount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到該紀錄"})
		return
	}

	tx, err := initializers.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "交易開始失敗"})
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	// 刪除紀錄
	_, err = tx.Exec("DELETE FROM records WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除紀錄失敗"})
		return
	}

	// 回滾帳戶餘額
	if recordType == "支出" {
		tx.Exec("UPDATE accounts SET balance = balance + ?, updated_at = ? WHERE id = ?", amount, now, accountID)
	} else {
		tx.Exec("UPDATE accounts SET balance = balance - ?, updated_at = ? WHERE id = ?", amount, now, accountID)
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "交易提交失敗"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "刪除成功"})
}

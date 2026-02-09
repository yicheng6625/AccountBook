package controllers

import (
	"accountbook/initializers"
	"accountbook/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetAccounts 取得所有帳戶列表
// 原因：前端帳戶頁與下拉選單需要完整帳戶資料
func GetAccounts(c *gin.Context) {
	rows, err := initializers.DB.Query("SELECT id, name, balance, sort_order, created_at, updated_at FROM accounts ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢帳戶失敗"})
		return
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(&a.ID, &a.Name, &a.Balance, &a.SortOrder, &a.CreatedAt, &a.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "讀取帳戶資料失敗"})
			return
		}
		accounts = append(accounts, a)
	}

	c.JSON(http.StatusOK, accounts)
}

// GetAccount 取得單一帳戶
func GetAccount(c *gin.Context) {
	id := c.Param("id")

	var a models.Account
	err := initializers.DB.QueryRow("SELECT id, name, balance, sort_order, created_at, updated_at FROM accounts WHERE id = ?", id).
		Scan(&a.ID, &a.Name, &a.Balance, &a.SortOrder, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到該帳戶"})
		return
	}

	c.JSON(http.StatusOK, a)
}

// CreateAccount 新增帳戶
// 原因：使用者可自訂帳戶（如新增電子錢包等）
func CreateAccount(c *gin.Context) {
	var input struct {
		Name    string `json:"name" binding:"required"`
		Balance float64 `json:"balance"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請提供帳戶名稱"})
		return
	}

	// 取得目前最大排序值，新帳戶排在最後
	var maxOrder int
	initializers.DB.QueryRow("SELECT COALESCE(MAX(sort_order), -1) FROM accounts").Scan(&maxOrder)

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := initializers.DB.Exec(
		"INSERT INTO accounts (name, balance, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		input.Name, input.Balance, maxOrder+1, now, now,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帳戶名稱已存在"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"id":      id,
		"name":    input.Name,
		"balance": input.Balance,
	})
}

// UpdateAccount 更新帳戶（名稱與餘額）
func UpdateAccount(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Name    *string  `json:"name"`
		Balance *float64 `json:"balance"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "輸入格式錯誤"})
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	// 依據有提供的欄位進行更新
	if input.Name != nil {
		initializers.DB.Exec("UPDATE accounts SET name = ?, updated_at = ? WHERE id = ?", *input.Name, now, id)
	}
	if input.Balance != nil {
		initializers.DB.Exec("UPDATE accounts SET balance = ?, updated_at = ? WHERE id = ?", *input.Balance, now, id)
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteAccount 刪除帳戶
// 原因：需檢查是否有關聯紀錄，有則不允許刪除
func DeleteAccount(c *gin.Context) {
	id := c.Param("id")

	// 檢查是否有關聯紀錄
	var count int
	initializers.DB.QueryRow("SELECT COUNT(*) FROM records WHERE account_id = ?", id).Scan(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "此帳戶尚有紀錄，無法刪除"})
		return
	}

	result, err := initializers.DB.Exec("DELETE FROM accounts WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除失敗"})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到該帳戶"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "刪除成功"})
}

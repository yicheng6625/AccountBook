package controllers

import (
	"accountbook/initializers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetStatistics 取得指定月份的分類統計
// 原因：統計頁需要各分類的金額與佔比，用於圓餅圖展示
func GetStatistics(c *gin.Context) {
	month := c.Query("month")
	year := c.Query("year")

	if month == "" && year == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請提供 month 或 year 參數"})
		return
	}

	var dateFilter string
	var param string
	if month != "" {
		dateFilter = "strftime('%Y-%m', r.date) = ?"
		param = month
	} else {
		dateFilter = "strftime('%Y', r.date) = ?"
		param = year
	}

	// 查詢各分類的支出統計
	rows, err := initializers.DB.Query(`
		SELECT c.id, c.name, r.type, SUM(r.amount) as total
		FROM records r
		JOIN categories c ON r.category_id = c.id
		WHERE `+dateFilter+`
		GROUP BY c.id, c.name, r.type
		ORDER BY total DESC
	`, param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢統計失敗"})
		return
	}
	defer rows.Close()

	type CategoryStat struct {
		ID         int     `json:"id"`
		Name       string  `json:"name"`
		Amount     float64 `json:"amount"`
		Percentage float64 `json:"percentage"`
	}

	var expenseCategories []CategoryStat
	var incomeCategories []CategoryStat
	var totalIncome, totalExpense float64

	for rows.Next() {
		var id int
		var name, recordType string
		var total float64
		if err := rows.Scan(&id, &name, &recordType, &total); err != nil {
			continue
		}
		stat := CategoryStat{ID: id, Name: name, Amount: total}
		if recordType == "收入" {
			totalIncome += total
			incomeCategories = append(incomeCategories, stat)
		} else {
			totalExpense += total
			expenseCategories = append(expenseCategories, stat)
		}
	}

	// 計算各分類的百分比
	for i := range expenseCategories {
		if totalExpense > 0 {
			expenseCategories[i].Percentage = expenseCategories[i].Amount / totalExpense * 100
		}
	}
	for i := range incomeCategories {
		if totalIncome > 0 {
			incomeCategories[i].Percentage = incomeCategories[i].Amount / totalIncome * 100
		}
	}

	// 決定回傳的時間標籤
	period := month
	if period == "" {
		period = year
	}

	c.JSON(http.StatusOK, gin.H{
		"period":             period,
		"total_income":       totalIncome,
		"total_expense":      totalExpense,
		"expense_categories": expenseCategories,
		"income_categories":  incomeCategories,
	})
}

// GetSummary 取得指定月份的收支總計
// 原因：前端統計頁頂部的總覽數字
func GetSummary(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請提供 month 參數"})
		return
	}

	var totalIncome, totalExpense float64

	initializers.DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) FROM records
		WHERE strftime('%Y-%m', date) = ? AND type = '收入'
	`, month).Scan(&totalIncome)

	initializers.DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) FROM records
		WHERE strftime('%Y-%m', date) = ? AND type = '支出'
	`, month).Scan(&totalExpense)

	c.JSON(http.StatusOK, gin.H{
		"month":         month,
		"total_income":  totalIncome,
		"total_expense": totalExpense,
		"balance":       totalIncome - totalExpense,
	})
}

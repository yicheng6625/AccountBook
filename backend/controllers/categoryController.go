package controllers

import (
	"accountbook/initializers"
	"accountbook/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetCategories 取得所有分類列表
// 原因：前端設定頁、下拉選單、Telegram Bot 都需要分類資料
func GetCategories(c *gin.Context) {
	rows, err := initializers.DB.Query("SELECT id, name, sort_order, created_at, updated_at FROM categories ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢分類失敗"})
		return
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.SortOrder, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "讀取分類資料失敗"})
			return
		}
		categories = append(categories, cat)
	}

	c.JSON(http.StatusOK, categories)
}

// CreateCategory 新增分類
// 原因：使用者可在設定頁自訂分類
func CreateCategory(c *gin.Context) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請提供分類名稱"})
		return
	}

	// 新分類排在最後
	var maxOrder int
	initializers.DB.QueryRow("SELECT COALESCE(MAX(sort_order), -1) FROM categories").Scan(&maxOrder)

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := initializers.DB.Exec(
		"INSERT INTO categories (name, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?)",
		input.Name, maxOrder+1, now, now,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "分類名稱已存在"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"id":   id,
		"name": input.Name,
	})
}

// UpdateCategory 更新分類名稱
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請提供分類名稱"})
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := initializers.DB.Exec("UPDATE categories SET name = ?, updated_at = ? WHERE id = ?", input.Name, now, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "分類名稱已存在"})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到該分類"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteCategory 刪除分類
// 原因：需檢查是否有關聯紀錄，有則不允許刪除
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	// 檢查是否有關聯紀錄
	var count int
	initializers.DB.QueryRow("SELECT COUNT(*) FROM records WHERE category_id = ?", id).Scan(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "此分類尚有紀錄，無法刪除"})
		return
	}

	result, err := initializers.DB.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除失敗"})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到該分類"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "刪除成功"})
}

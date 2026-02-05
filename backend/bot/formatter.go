package bot

import (
	"accountbook/initializers"
	"accountbook/models"
	"fmt"
	"strings"
)

// FormatSuccess 格式化新增成功的回覆訊息
// 原因：依照 README 定義的回覆格式，讓使用者確認紀錄內容
func FormatSuccess(date, accountName, recordType string, amount float64, item, categoryName, note string) string {
	return fmt.Sprintf(`新增成功！紀錄如下：
時間：%s
帳戶名稱：%s
類型：%s
金額：%.0f
項目名稱：%s
分類：%s
備註：%s`, date, accountName, recordType, amount, item, categoryName, note)
}

// FormatError 格式化新增失敗的回覆訊息
// 原因：動態列出帳戶與分類，幫助使用者輸入正確格式
func FormatError() string {
	accounts := getAccountList()
	categories := getCategoryList()

	return fmt.Sprintf(`新增失敗！請確認格式是否正確：
時間(昨天/今天/明天/日期[可接受"2026/01/01"或"01/01"或"1/1"])
帳戶名稱(可省略，預設為現金%s)
收入/支出(可省略，預設為支出)
支出金額
項目名稱
分類(%s)
備註(可省略)`, accounts, categories)
}

// FormatUsage 格式化新增格式說明
// 原因：回應 /新增格式 指令
func FormatUsage() string {
	accounts := getAccountList()
	categories := getCategoryList()

	return fmt.Sprintf(`新增紀錄格式如下：
時間(昨天/今天/明天/日期[可接受"2026/01/01"或"01/01"或"1/1"])
帳戶名稱(可省略，預設為現金%s)
收入/支出(可省略，預設為支出)
支出金額
項目名稱
分類(%s)
備註(可省略)`, accounts, categories)
}

// FormatCategories 格式化分類列表
// 原因：回應 /查詢分類 指令，列出編號與名稱
func FormatCategories() string {
	rows, err := initializers.DB.Query("SELECT id, name FROM categories ORDER BY sort_order")
	if err != nil {
		return "查詢分類失敗"
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("%d: %s", cat.ID, cat.Name))
	}

	return strings.Join(lines, "\n")
}

// FormatAccounts 格式化帳戶列表
// 原因：回應 /查詢帳戶 指令，列出名稱與餘額
func FormatAccounts() string {
	rows, err := initializers.DB.Query("SELECT name, balance FROM accounts ORDER BY sort_order")
	if err != nil {
		return "查詢帳戶失敗"
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var name string
		var balance float64
		if err := rows.Scan(&name, &balance); err != nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %.0f", name, balance))
	}

	return strings.Join(lines, "\n")
}

// getAccountList 取得帳戶清單字串（供格式說明使用）
func getAccountList() string {
	rows, err := initializers.DB.Query("SELECT id, name FROM accounts ORDER BY sort_order")
	if err != nil {
		return ""
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		items = append(items, fmt.Sprintf("%d:%s", id, name))
	}

	if len(items) == 0 {
		return ""
	}
	return " {" + strings.Join(items, ", ") + "}"
}

// getCategoryList 取得分類清單字串（供格式說明使用）
func getCategoryList() string {
	rows, err := initializers.DB.Query("SELECT id, name FROM categories ORDER BY sort_order")
	if err != nil {
		return ""
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		items = append(items, fmt.Sprintf("%d:%s", id, name))
	}

	if len(items) == 0 {
		return ""
	}
	return strings.Join(items, ", ")
}

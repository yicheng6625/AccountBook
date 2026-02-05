package bot

import (
	"accountbook/initializers"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParsedRecord 解析後的紀錄資料
type ParsedRecord struct {
	Date       string
	AccountID  int
	Type       string
	Amount     float64
	Item       string
	CategoryID int
	Note       string
}

// ParseRecord 解析多行訊息為紀錄資料
// 原因：Telegram Bot 的核心邏輯，需處理多種省略格式
// 格式（完整版）：
//
//	時間 / 帳戶名稱(可省略) / 收入或支出(可省略) / 金額 / 項目名稱 / 分類 / 備註(可省略)
func ParseRecord(message string) (*ParsedRecord, error) {
	lines := splitLines(message)
	if len(lines) < 3 {
		return nil, fmt.Errorf("格式錯誤：至少需要 3 行（時間、金額、項目名稱+分類）")
	}

	// 第 1 行固定為時間
	date, err := parseDate(lines[0])
	if err != nil {
		return nil, fmt.Errorf("時間格式錯誤：%s", lines[0])
	}

	remaining := lines[1:]
	record := &ParsedRecord{
		Date:      date,
		AccountID: getDefaultAccountID(), // 預設帳戶：現金
		Type:      "支出",                  // 預設類型：支出
	}

	// 依據剩餘行數判斷各欄位位置
	switch len(remaining) {
	case 2:
		// 最簡格式：金額 / 項目名稱
		// 注意：此格式缺少分類，無法使用
		return nil, fmt.Errorf("格式錯誤：至少需要提供分類")

	case 3:
		// 省略帳戶、類型、備註：金額 / 項目名稱 / 分類
		amount, err := parseAmount(remaining[0])
		if err != nil {
			return nil, err
		}
		categoryID, err := parseCategoryID(remaining[2])
		if err != nil {
			return nil, err
		}
		record.Amount = amount
		record.Item = remaining[1]
		record.CategoryID = categoryID

	case 4:
		// 可能是以下兩種格式之一：
		// A: 帳戶 / 金額 / 項目 / 分類
		// B: 收入或支出 / 金額 / 項目 / 分類
		if isAccountName(remaining[0]) {
			// 格式 A
			accountID, _ := resolveAccountID(remaining[0])
			amount, err := parseAmount(remaining[1])
			if err != nil {
				return nil, err
			}
			categoryID, err := parseCategoryID(remaining[3])
			if err != nil {
				return nil, err
			}
			record.AccountID = accountID
			record.Amount = amount
			record.Item = remaining[2]
			record.CategoryID = categoryID
		} else if isType(remaining[0]) {
			// 格式 B
			amount, err := parseAmount(remaining[1])
			if err != nil {
				return nil, err
			}
			categoryID, err := parseCategoryID(remaining[3])
			if err != nil {
				return nil, err
			}
			record.Type = normalizeType(remaining[0])
			record.Amount = amount
			record.Item = remaining[2]
			record.CategoryID = categoryID
		} else {
			// 嘗試當作 金額 / 項目 / 分類 / 備註
			amount, err := parseAmount(remaining[0])
			if err != nil {
				return nil, fmt.Errorf("無法識別格式，第二行應為帳戶名稱、收入/支出、或金額")
			}
			categoryID, err := parseCategoryID(remaining[2])
			if err != nil {
				return nil, err
			}
			record.Amount = amount
			record.Item = remaining[1]
			record.CategoryID = categoryID
			record.Note = remaining[3]
		}

	case 5:
		// 帳戶 / 類型 / 金額 / 項目 / 分類
		// 或 帳戶 / 金額 / 項目 / 分類 / 備註
		if isType(remaining[1]) {
			accountID, err := resolveAccountID(remaining[0])
			if err != nil {
				return nil, err
			}
			amount, err := parseAmount(remaining[2])
			if err != nil {
				return nil, err
			}
			categoryID, err := parseCategoryID(remaining[4])
			if err != nil {
				return nil, err
			}
			record.AccountID = accountID
			record.Type = normalizeType(remaining[1])
			record.Amount = amount
			record.Item = remaining[3]
			record.CategoryID = categoryID
		} else {
			accountID, err := resolveAccountID(remaining[0])
			if err != nil {
				return nil, err
			}
			amount, err := parseAmount(remaining[1])
			if err != nil {
				return nil, err
			}
			categoryID, err := parseCategoryID(remaining[3])
			if err != nil {
				return nil, err
			}
			record.AccountID = accountID
			record.Amount = amount
			record.Item = remaining[2]
			record.CategoryID = categoryID
			record.Note = remaining[4]
		}

	case 6:
		// 完整格式：帳戶 / 類型 / 金額 / 項目 / 分類 / 備註
		accountID, err := resolveAccountID(remaining[0])
		if err != nil {
			return nil, err
		}
		amount, err := parseAmount(remaining[2])
		if err != nil {
			return nil, err
		}
		categoryID, err := parseCategoryID(remaining[4])
		if err != nil {
			return nil, err
		}
		record.AccountID = accountID
		record.Type = normalizeType(remaining[1])
		record.Amount = amount
		record.Item = remaining[3]
		record.CategoryID = categoryID
		record.Note = remaining[5]

	default:
		return nil, fmt.Errorf("格式錯誤：行數過多")
	}

	return record, nil
}

// splitLines 分割訊息為行並去除空白
func splitLines(message string) []string {
	raw := strings.Split(message, "\n")
	var lines []string
	for _, line := range raw {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

// parseDate 解析日期字串
// 原因：支援多種日期格式（昨天/今天/明天/完整日期/短日期）
func parseDate(input string) (string, error) {
	now := time.Now()

	switch input {
	case "昨天":
		return now.AddDate(0, 0, -1).Format("2006-01-02"), nil
	case "今天":
		return now.Format("2006-01-02"), nil
	case "明天":
		return now.AddDate(0, 0, 1).Format("2006-01-02"), nil
	}

	// 嘗試完整日期格式：2026/01/01
	formats := []string{"2006/01/02", "2006/1/2"}
	for _, f := range formats {
		if t, err := time.Parse(f, input); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}

	// 嘗試短日期格式：01/01 或 1/1（使用當前年份）
	shortFormats := []string{"01/02", "1/2"}
	for _, f := range shortFormats {
		if t, err := time.Parse(f, input); err == nil {
			result := time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
			return result.Format("2006-01-02"), nil
		}
	}

	return "", fmt.Errorf("無法解析日期: %s", input)
}

// parseAmount 解析金額
func parseAmount(input string) (float64, error) {
	amount, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("金額格式錯誤：%s", input)
	}
	if amount <= 0 {
		return 0, fmt.Errorf("金額必須大於 0")
	}
	return amount, nil
}

// parseCategoryID 解析分類（支援編號或名稱）
// 原因：使用者可以輸入分類編號（快速）或分類名稱（直覺）
func parseCategoryID(input string) (int, error) {
	// 嘗試作為編號
	if id, err := strconv.Atoi(input); err == nil {
		var exists int
		err := initializers.DB.QueryRow("SELECT COUNT(*) FROM categories WHERE id = ?", id).Scan(&exists)
		if err == nil && exists > 0 {
			return id, nil
		}
	}

	// 嘗試作為名稱
	var id int
	err := initializers.DB.QueryRow("SELECT id FROM categories WHERE name = ?", input).Scan(&id)
	if err == nil {
		return id, nil
	}

	return 0, fmt.Errorf("找不到分類：%s", input)
}

// isAccountName 判斷文字是否為帳戶名稱或編號
func isAccountName(input string) bool {
	// 嘗試作為編號
	if id, err := strconv.Atoi(input); err == nil {
		var exists int
		initializers.DB.QueryRow("SELECT COUNT(*) FROM accounts WHERE id = ?", id).Scan(&exists)
		if exists > 0 {
			return true
		}
	}

	// 嘗試作為名稱
	var exists int
	initializers.DB.QueryRow("SELECT COUNT(*) FROM accounts WHERE name = ?", input).Scan(&exists)
	return exists > 0
}

// resolveAccountID 取得帳戶 ID（支援名稱或編號）
func resolveAccountID(input string) (int, error) {
	// 嘗試作為編號
	if id, err := strconv.Atoi(input); err == nil {
		var exists int
		initializers.DB.QueryRow("SELECT COUNT(*) FROM accounts WHERE id = ?", id).Scan(&exists)
		if exists > 0 {
			return id, nil
		}
	}

	// 嘗試作為名稱
	var id int
	err := initializers.DB.QueryRow("SELECT id FROM accounts WHERE name = ?", input).Scan(&id)
	if err == nil {
		return id, nil
	}

	return 0, fmt.Errorf("找不到帳戶：%s", input)
}

// isType 判斷文字是否為收入/支出
func isType(input string) bool {
	return input == "收入" || input == "支出"
}

// normalizeType 標準化類型文字
func normalizeType(input string) string {
	if input == "收入" {
		return "收入"
	}
	return "支出"
}

// getDefaultAccountID 取得預設帳戶（現金）的 ID
// 原因：省略帳戶欄位時使用預設值
func getDefaultAccountID() int {
	var id int
	err := initializers.DB.QueryRow("SELECT id FROM accounts WHERE name = '現金' LIMIT 1").Scan(&id)
	if err != nil {
		return 1 // 若找不到現金帳戶，預設使用 ID 1
	}
	return id
}

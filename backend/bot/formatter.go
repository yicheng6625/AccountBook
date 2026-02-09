package bot

import (
	"accountbook/initializers"
	"accountbook/models"
	"accountbook/services"
	"fmt"
	"strings"
)

// FormatSuccess 格式化新增成功的回覆訊息
func FormatSuccess(date, accountName, recordType string, amount float64, item, categoryName, note string) string {
	return fmt.Sprintf(`✅ 新增成功！

📅 %s
💰 %s %.0f
📝 %s
🏷 %s
🏦 %s
📌 %s`, date, recordType, amount, item, categoryName, accountName, note)
}

// FormatPreview 格式化新增紀錄的預覽訊息
// 原因：顯示目前所有欄位值，讓使用者一目了然，點擊按鈕即可修改
func FormatPreview(s *Session) string {
	accountName := resolveAccountName(s.AccountID)
	categoryName := resolveCategoryName(s.CategoryID)

	amountStr := "（未填）"
	if s.Amount > 0 {
		amountStr = fmt.Sprintf("%.0f", s.Amount)
	}

	itemStr := s.Item
	if itemStr == "" {
		itemStr = "（未填）"
	}

	noteStr := s.Note
	if noteStr == "" {
		noteStr = "（無）"
	}

	return fmt.Sprintf(`📋 新增紀錄

📅 日期：%s
🏦 帳戶：%s
💱 類型：%s
💰 金額：%s
📝 項目：%s
🏷 分類：%s
📌 備註：%s

點擊下方按鈕修改欄位，或按「✅ 確認送出」`, s.Date, accountName, s.Type, amountStr, itemStr, categoryName, noteStr)
}

// BuildPreviewKeyboard 建立預覽訊息的 Inline Keyboard
// 原因：每個欄位一個按鈕，使用者點擊後進入該欄位的編輯狀態
func BuildPreviewKeyboard(s *Session) services.InlineKeyboardMarkup {
	// 第一排：日期、帳戶
	row1 := []services.InlineKeyboardButton{
		{Text: "📅 日期", CallbackData: "edit_date"},
		{Text: "🏦 帳戶", CallbackData: "edit_account"},
	}

	// 第二排：類型、金額
	row2 := []services.InlineKeyboardButton{
		{Text: "💱 類型", CallbackData: "edit_type"},
		{Text: "💰 金額", CallbackData: "edit_amount"},
	}

	// 第三排：項目、分類
	row3 := []services.InlineKeyboardButton{
		{Text: "📝 項目", CallbackData: "edit_item"},
		{Text: "🏷 分類", CallbackData: "edit_category"},
	}

	// 第四排：備註
	row4 := []services.InlineKeyboardButton{
		{Text: "📌 備註", CallbackData: "edit_note"},
	}

	// 第五排：確認送出、取消
	row5 := []services.InlineKeyboardButton{
		{Text: "✅ 確認送出", CallbackData: "confirm"},
		{Text: "❌ 取消", CallbackData: "cancel"},
	}

	return services.InlineKeyboardMarkup{
		InlineKeyboard: [][]services.InlineKeyboardButton{row1, row2, row3, row4, row5},
	}
}

// BuildAccountKeyboard 建立帳戶選擇的 Inline Keyboard
// 原因：列出所有帳戶讓使用者直接點擊選擇，不需要手動輸入
func BuildAccountKeyboard() services.InlineKeyboardMarkup {
	rows, _ := initializers.DB.Query("SELECT id, name FROM accounts ORDER BY sort_order")
	defer rows.Close()

	var buttons [][]services.InlineKeyboardButton
	var row []services.InlineKeyboardButton

	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		row = append(row, services.InlineKeyboardButton{
			Text:         name,
			CallbackData: fmt.Sprintf("set_account_%d", id),
		})
		// 每排 2 個按鈕
		if len(row) == 2 {
			buttons = append(buttons, row)
			row = nil
		}
	}
	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	return services.InlineKeyboardMarkup{InlineKeyboard: buttons}
}

// BuildTypeKeyboard 建立收入/支出選擇的 Inline Keyboard
func BuildTypeKeyboard() services.InlineKeyboardMarkup {
	return services.InlineKeyboardMarkup{
		InlineKeyboard: [][]services.InlineKeyboardButton{
			{
				{Text: "💸 支出", CallbackData: "set_type_支出"},
				{Text: "💵 收入", CallbackData: "set_type_收入"},
			},
		},
	}
}

// BuildCategoryKeyboard 建立分類選擇的 Inline Keyboard
// 原因：列出所有分類讓使用者直接點擊選擇
func BuildCategoryKeyboard() services.InlineKeyboardMarkup {
	rows, _ := initializers.DB.Query("SELECT id, name FROM categories ORDER BY sort_order")
	defer rows.Close()

	var buttons [][]services.InlineKeyboardButton
	var row []services.InlineKeyboardButton

	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		row = append(row, services.InlineKeyboardButton{
			Text:         name,
			CallbackData: fmt.Sprintf("set_category_%d", id),
		})
		// 每排 3 個按鈕
		if len(row) == 3 {
			buttons = append(buttons, row)
			row = nil
		}
	}
	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	return services.InlineKeyboardMarkup{InlineKeyboard: buttons}
}

// BuildDateKeyboard 建立日期快捷選擇的 Inline Keyboard
func BuildDateKeyboard() services.InlineKeyboardMarkup {
	return services.InlineKeyboardMarkup{
		InlineKeyboard: [][]services.InlineKeyboardButton{
			{
				{Text: "前天", CallbackData: "set_date_-2"},
				{Text: "昨天", CallbackData: "set_date_-1"},
				{Text: "今天", CallbackData: "set_date_0"},
				{Text: "明天", CallbackData: "set_date_1"},
			},
		},
	}
}

// resolveAccountName 取得帳戶名稱
func resolveAccountName(id int) string {
	var name string
	err := initializers.DB.QueryRow("SELECT name FROM accounts WHERE id = ?", id).Scan(&name)
	if err != nil {
		return "未知"
	}
	return name
}

// resolveCategoryName 取得分類名稱
func resolveCategoryName(id int) string {
	var name string
	err := initializers.DB.QueryRow("SELECT name FROM categories WHERE id = ?", id).Scan(&name)
	if err != nil {
		return "未知"
	}
	return name
}

// === 轉帳格式化 ===

// FormatTransferPreview 格式化轉帳預覽訊息
func FormatTransferPreview(s *Session) string {
	fromName := resolveAccountName(s.AccountID)
	toName := resolveAccountName(s.ToAccountID)

	amountStr := "（未填）"
	if s.Amount > 0 {
		amountStr = fmt.Sprintf("%.0f", s.Amount)
	}

	noteStr := s.Note
	if noteStr == "" {
		noteStr = "（無）"
	}

	return fmt.Sprintf(`🔄 轉帳

🏦 轉出：%s
🏦 轉入：%s
💰 金額：%s
📌 備註：%s

點擊下方按鈕修改欄位，或按「✅ 確認轉帳」`, fromName, toName, amountStr, noteStr)
}

// BuildTransferKeyboard 建立轉帳預覽的 Inline Keyboard
func BuildTransferKeyboard(s *Session) services.InlineKeyboardMarkup {
	row1 := []services.InlineKeyboardButton{
		{Text: "🏦 轉出帳戶", CallbackData: "transfer_edit_from"},
		{Text: "🏦 轉入帳戶", CallbackData: "transfer_edit_to"},
	}
	row2 := []services.InlineKeyboardButton{
		{Text: "💰 金額", CallbackData: "transfer_edit_amount"},
		{Text: "📌 備註", CallbackData: "transfer_edit_note"},
	}
	row3 := []services.InlineKeyboardButton{
		{Text: "✅ 確認轉帳", CallbackData: "transfer_confirm"},
		{Text: "❌ 取消", CallbackData: "cancel"},
	}

	return services.InlineKeyboardMarkup{
		InlineKeyboard: [][]services.InlineKeyboardButton{row1, row2, row3},
	}
}

// BuildTransferAccountKeyboard 建立轉帳用帳戶選擇鍵盤
func BuildTransferAccountKeyboard(prefix string) services.InlineKeyboardMarkup {
	rows, _ := initializers.DB.Query("SELECT id, name FROM accounts ORDER BY sort_order")
	defer rows.Close()

	var buttons [][]services.InlineKeyboardButton
	var row []services.InlineKeyboardButton

	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		row = append(row, services.InlineKeyboardButton{
			Text:         name,
			CallbackData: fmt.Sprintf("%s%d", prefix, id),
		})
		if len(row) == 2 {
			buttons = append(buttons, row)
			row = nil
		}
	}
	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	return services.InlineKeyboardMarkup{InlineKeyboard: buttons}
}

// FormatTransferSuccess 格式化轉帳成功訊息
func FormatTransferSuccess(fromName, toName string, amount float64, note string) string {
	noteStr := note
	if noteStr == "" {
		noteStr = "（無）"
	}
	return fmt.Sprintf(`✅ 轉帳成功！

🏦 %s ➡️ %s
💰 %.0f
📌 %s`, fromName, toName, amount, noteStr)
}

// FormatRecentRecords 格式化最近紀錄的查詢結果（分頁顯示）
func FormatRecentRecords(offset, pageSize int) (string, int) {
	// 先查詢總筆數
	var total int
	initializers.DB.QueryRow("SELECT COUNT(*) FROM records").Scan(&total)

	if total == 0 {
		return "📭 目前沒有任何紀錄", 0
	}

	rows, err := initializers.DB.Query(`
		SELECT r.date, a.name, r.type, r.amount, r.item, c.name, r.note
		FROM records r
		LEFT JOIN accounts a ON r.account_id = a.id
		LEFT JOIN categories c ON r.category_id = c.id
		ORDER BY r.date DESC, r.id DESC
		LIMIT ? OFFSET ?
	`, pageSize, offset)
	if err != nil {
		return "查詢紀錄失敗", 0
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var date, accountName, recordType, item, categoryName, note string
		var amount float64
		rows.Scan(&date, &accountName, &recordType, &amount, &item, &categoryName, &note)

		line := fmt.Sprintf("📅 %s｜%s %.0f\n📝 %s｜🏷 %s｜🏦 %s",
			date, recordType, amount, item, categoryName, accountName)
		if note != "" {
			line += fmt.Sprintf("\n📌 %s", note)
		}
		lines = append(lines, line)
	}

	page := offset/pageSize + 1
	totalPages := (total + pageSize - 1) / pageSize
	header := fmt.Sprintf("📋 最近紀錄（第 %d/%d 頁）\n", page, totalPages)

	return header + "\n" + strings.Join(lines, "\n\n"), total
}

// BuildPaginationKeyboard 建立翻頁按鈕
func BuildPaginationKeyboard(offset, pageSize, total int) services.InlineKeyboardMarkup {
	var row []services.InlineKeyboardButton

	if offset > 0 {
		row = append(row, services.InlineKeyboardButton{
			Text:         "⬅️ 上一頁",
			CallbackData: fmt.Sprintf("recent_page_%d", offset-pageSize),
		})
	}

	if offset+pageSize < total {
		row = append(row, services.InlineKeyboardButton{
			Text:         "➡️ 下一頁",
			CallbackData: fmt.Sprintf("recent_page_%d", offset+pageSize),
		})
	}

	var buttons [][]services.InlineKeyboardButton
	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	return services.InlineKeyboardMarkup{InlineKeyboard: buttons}
}

// === 以下保留原有的查詢格式化功能 ===

// FormatCategories 格式化分類列表
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

// FormatUsage 格式化使用說明
func FormatUsage() string {
	return `📋 記帳機器人使用說明

輸入 /new 即可開始新增紀錄
所有欄位都已預設好，只需修改需要的項目

指令列表：
/new - 開始記帳
/transfer - 帳戶轉帳
/recent - 查看最近紀錄
/start - 顯示此說明
/查詢分類 - 查看所有分類
/查詢帳戶 - 查看所有帳戶餘額
/cancel - 取消目前操作`
}

// FormatError 格式化錯誤訊息（保留給多行格式解析失敗時使用）
func FormatError() string {
	return "輸入格式有誤，請重新輸入"
}

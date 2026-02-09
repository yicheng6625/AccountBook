package initializers

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // 純 Go SQLite 驅動，不需要 CGO
)

// DB 全域資料庫連線實例
var DB *sql.DB

// InitDB 初始化 SQLite 資料庫連線與資料表
// 原因：程式啟動時建立連線，並確保資料表結構存在
func InitDB(dbPath string) {
	var err error
	DB, err = sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		log.Fatalf("無法開啟資料庫: %v", err)
	}

	// 驗證連線是否正常
	if err = DB.Ping(); err != nil {
		log.Fatalf("資料庫連線失敗: %v", err)
	}

	// 建立資料表結構
	createTables()

	// 插入預設資料
	insertDefaults()

	log.Println("資料庫初始化完成")
}

// createTables 建立所有資料表
// 原因：使用 IF NOT EXISTS 確保重複執行不會報錯
func createTables() {
	statements := []string{
		// 帳戶資料表
		`CREATE TABLE IF NOT EXISTS accounts (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT    NOT NULL UNIQUE,
			balance     REAL    NOT NULL DEFAULT 0,
			sort_order  INTEGER NOT NULL DEFAULT 0,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 分類資料表
		`CREATE TABLE IF NOT EXISTS categories (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT    NOT NULL UNIQUE,
			sort_order  INTEGER NOT NULL DEFAULT 0,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 記帳紀錄資料表
		`CREATE TABLE IF NOT EXISTS records (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			date        DATE    NOT NULL,
			account_id  INTEGER NOT NULL,
			type        TEXT    NOT NULL DEFAULT '支出',
			amount      REAL    NOT NULL,
			item        TEXT    NOT NULL,
			category_id INTEGER NOT NULL,
			note        TEXT    DEFAULT '',
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (account_id)  REFERENCES accounts(id),
			FOREIGN KEY (category_id) REFERENCES categories(id)
		)`,

		// 索引：加速日期查詢（首頁行事曆最常用）
		`CREATE INDEX IF NOT EXISTS idx_records_date ON records(date)`,

		// 索引：加速分類統計
		`CREATE INDEX IF NOT EXISTS idx_records_category ON records(category_id)`,

		// 索引：加速帳戶查詢
		`CREATE INDEX IF NOT EXISTS idx_records_account ON records(account_id)`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			log.Fatalf("建立資料表失敗: %v\nSQL: %s", err, stmt)
		}
	}
}

// insertDefaults 插入預設資料
// 原因：首次啟動時提供預設帳戶與分類，避免空白系統
// 若已有資料則不插入，避免重啟時覆蓋使用者自訂資料
func insertDefaults() {
	// 若已有帳戶則跳過
	var accountCount int
	DB.QueryRow("SELECT COUNT(*) FROM accounts").Scan(&accountCount)
	if accountCount == 0 {
		accounts := []struct {
			name      string
			sortOrder int
		}{
			{"現金", 0},
			{"信用卡", 1},
			{"銀行帳戶", 2},
		}
		for _, a := range accounts {
			DB.Exec("INSERT INTO accounts (name, balance, sort_order) VALUES (?, 0, ?)", a.name, a.sortOrder)
		}
	}

	// 若已有分類則跳過
	var categoryCount int
	DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&categoryCount)
	if categoryCount == 0 {
		categories := []struct {
			name      string
			sortOrder int
		}{
			{"飲食", 0},
			{"交通", 1},
			{"服飾", 2},
			{"3C", 3},
			{"娛樂", 4},
			{"其他", 5},
		}
		for _, c := range categories {
			DB.Exec("INSERT INTO categories (name, sort_order) VALUES (?, ?)", c.name, c.sortOrder)
		}
	}
}

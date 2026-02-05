package initializers

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv 載入環境變數
// 原因：優先讀取 .env 檔案，若不存在則使用系統環境變數（Docker 環境）
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("未找到 .env 檔案，使用系統環境變數")
	}
}

// GetEnv 取得環境變數，若不存在則回傳預設值
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

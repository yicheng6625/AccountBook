package main

import (
	"accountbook/bot"
	"accountbook/controllers"
	"accountbook/initializers"
	"accountbook/services"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	// 載入環境變數
	initializers.LoadEnv()

	// 初始化資料庫
	dbPath := initializers.GetEnv("DB_PATH", "./data/accountbook.db")
	initializers.InitDB(dbPath)
}

func main() {
	r := gin.Default()

	// 設定 CORS，允許前端跨域呼叫
	r.Use(cors.Default())

	api := r.Group("/api")
	{
		// 健康檢查
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		// 紀錄相關路由
		api.GET("/records", controllers.GetRecords)
		api.GET("/records/:id", controllers.GetRecord)
		api.POST("/records", controllers.CreateRecord)
		api.PUT("/records/:id", controllers.UpdateRecord)
		api.DELETE("/records/:id", controllers.DeleteRecord)

		// 帳戶相關路由
		api.GET("/accounts", controllers.GetAccounts)
		api.GET("/accounts/:id", controllers.GetAccount)
		api.POST("/accounts", controllers.CreateAccount)
		api.PUT("/accounts/:id", controllers.UpdateAccount)
		api.DELETE("/accounts/:id", controllers.DeleteAccount)

		// 分類相關路由
		api.GET("/categories", controllers.GetCategories)
		api.POST("/categories", controllers.CreateCategory)
		api.PUT("/categories/:id", controllers.UpdateCategory)
		api.DELETE("/categories/:id", controllers.DeleteCategory)

		// 轉帳路由
		api.POST("/transfer", controllers.CreateTransfer)

		// 統計相關路由
		api.GET("/statistics", controllers.GetStatistics)
		api.GET("/statistics/summary", controllers.GetSummary)

		// Telegram Webhook
		api.POST("/telegram/webhook", bot.HandleWebhook)
	}

	// 啟動 Telegram Bot Webhook（若有設定 token）
	token := initializers.GetEnv("TELEGRAM_BOT_TOKEN", "")
	webhookURL := initializers.GetEnv("TELEGRAM_WEBHOOK_URL", "")
	if token != "" && webhookURL != "" {
		services.SetupWebhook(token, webhookURL)
	}

	// 啟動伺服器
	port := initializers.GetEnv("PORT", "8080")
	log.Printf("伺服器啟動於 :%s", port)
	r.Run(":" + port)
}

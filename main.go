package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/JohnsonYuanTW/NCAEats/config"
	"github.com/JohnsonYuanTW/NCAEats/handler"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func initDBConn(config *config.Config, log *logrus.Logger) (*gorm.DB, error) {
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(config.DBUsername, config.DBPassword),
		Host:     fmt.Sprintf("%s:%s", config.DBURL, config.DBPort),
		Path:     config.DBName,
		RawQuery: (&url.Values{"sslmode": []string{"disable"}, "TimeZone": []string{"Asia/Taipei"}}).Encode(),
	}

	dsn := u.String()

	// Open DB connection
	newLogger := logger.New(log, logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL database instance: %v", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(config.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(config.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.DBConnMaxLifetime)

	return db, nil
}

func main() {
	// Create logger
	log := logrus.New()

	// Create Template Handler
	templates, err := handler.NewTemplateHandler("./templates")
	if err != nil {
		log.WithError(err).Fatal("JSON 模板初始化失敗")
	}
	log.Info("JSON 模板初始化成功")

	// Load environment variables
	s, err := config.LoadEnvVariables()
	if err != nil {
		log.WithError(err).Fatal("無法載入環境變數")
	}
	log.Info("環境變數載入成功")

	// Initialize DB conn and models
	db, err := initDBConn(s, log)
	if err != nil {
		log.WithError(err).Fatal("資料庫連線失敗")
	}
	log.Info("資料庫連線成功")

	// Create LineBot client
	bot, err := linebot.New(s.ChannelSecret, s.ChannelAccessToken)
	if err != nil {
		log.WithError(err).Fatal("Linebot 建立失敗")
	}
	log.Info("Linebot 建立成功")

	// Create appHandler
	appHandler, err := handler.NewAppHandler(log, templates, s, bot, db)
	if err != nil {
		log.WithError(err).Fatal("模型初始化失敗")
	}
	log.Info("模型初始化成功")

	log.Info("程式已啟動...")

	// Set up routes
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), customLogger(log))
	r.POST("/callback", appHandler.CallbackHandler)
	r.GET("/userReport/:reportID", func(c *gin.Context) {
		reportID := c.Params.ByName("reportID")

		// Look up the orderID of ReportID in the db
		orderID, err := appHandler.OrderRepo.GetOrderIDByReportID(reportID)
		if err != nil {
			c.String(http.StatusNotFound, "Report not found")
			log.WithError(err).Errorf("無法取得 %s 報表對應的訂單", reportID)
			return
		}

		// Get reportHTML
		reportHTML, err := appHandler.OrderRepo.GetOrderReportByOrderID(orderID)
		if err != nil {
			c.String(http.StatusNotFound, "Report not found")
			log.WithError(err).Errorf("無法取得 %s 報表", reportID)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(reportHTML))
	})

	// Start server
	addr := fmt.Sprintf(":%s", s.Port)
	if err := r.RunTLS(addr, s.SSLCertfilePath, s.SSLKeyPath); err != nil {
		log.WithError(err).Fatal("無法啟動網頁伺服器")
	}
}

func customLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Log only 4xx and 5xx responses, excluding favicon.ico
		if (c.Writer.Status() >= 400 && c.Writer.Status() < 600) && c.Request.URL.Path != "/favicon.ico" {
			endTime := time.Now()
			latency := endTime.Sub(startTime)
			clientIP := c.ClientIP()
			method := c.Request.Method
			statusCode := c.Writer.Status()

			entry := log.WithFields(logrus.Fields{
				"status_code": statusCode,
				"latency":     latency,
				"client_ip":   clientIP,
				"method":      method,
				"path":        c.Request.URL.Path,
			})

			if statusCode >= 500 {
				entry.Error("[GIN] Internal Server Error")
			} else {
				entry.Warn("[GIN] Client Error")
			}
		}
	}
}

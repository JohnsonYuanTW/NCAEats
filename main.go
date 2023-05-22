package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/JohnsonYuanTW/NCAEats/config"
	"github.com/JohnsonYuanTW/NCAEats/handler"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func initDBConn(env map[string]string) (*gorm.DB, error) {
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(url.QueryEscape(env["DB_USERNAME"]), url.QueryEscape(env["DB_PASSWORD"])),
		Host:     fmt.Sprintf("%s:%s", env["DB_URL"], env["DB_PORT"]),
		Path:     url.QueryEscape(env["DB_NAME"]),
		RawQuery: (&url.Values{"sslmode": []string{"disable"}, "TimeZone": []string{"Asia/Taipei"}}).Encode(),
	}

	dsn := u.String()

	// Open DB connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL database instance: %v", err)
	}

	// Set connection pool parameters
	maxIdleConns, err := strconv.Atoi(env["DB_MAX_IDLE_CONNS"])
	if err != nil {
		return nil, fmt.Errorf("invalid value for max idle connections: %v", err)
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)

	maxOpenConns, err := strconv.Atoi(env["DB_MAX_OPEN_CONNS"])
	if err != nil {
		return nil, fmt.Errorf("invalid value for max open connections: %v", err)
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)

	connMaxLifetime, err := time.ParseDuration(env["DB_CONN_MAX_LIFETIME"])
	if err != nil {
		return nil, fmt.Errorf("invalid value for max connection lifetime: %v", err)
	}
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

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
	env, err := config.LoadEnvVariables()
	if err != nil {
		log.WithError(err).Fatal("無法載入環境變數")
	}
	log.Info("環境變數載入成功")

	// Initialize DB conn and models
	db, err := initDBConn(env)
	if err != nil {
		log.WithError(err).Fatal("資料庫連線失敗")
	}
	log.Info("資料庫連線成功")

	// Create LineBot client
	bot, err := linebot.New(env["ChannelSecret"], env["ChannelAccessToken"])
	if err != nil {
		log.WithError(err).Fatal("Linebot 建立失敗")
	}
	log.Info("Linebot 建立成功")

	// Create appHandler
	appHandler, err := handler.NewAppHandler(log, templates, env, bot, db)
	if err != nil {
		log.WithError(err).Fatal("模型初始化失敗")
	}
	log.Info("模型初始化成功")

	log.Info("程式已啟動...")

	// Set up routes
	http.HandleFunc("/callback", appHandler.CallbackHandler)
	http.HandleFunc("/userReport/", func(w http.ResponseWriter, r *http.Request) {
		// Get the report ID from the URL
		reportID := strings.TrimPrefix(r.URL.Path, "/userReport/")

		// Look up the orderID of ReportID in the db
		orderID, err := appHandler.OrderRepo.GetOrderIDByReportID(reportID)
		if err != nil {
			http.Error(w, "Report not found", http.StatusNotFound)
			log.WithError(err).Errorf("無法取得 %s 報表對應的訂單", reportID)
		}

		// Get reportHTML
		reportHTML, err := appHandler.OrderRepo.GetOrderReportByOrderID(orderID)
		if err != nil {
			http.Error(w, "Report not found", http.StatusNotFound)
			log.WithError(err).Errorf("無法取得 %s 報表", reportID)
		}

		// Write to HTML
		fmt.Fprint(w, reportHTML)
	})

	// Start server
	addr := fmt.Sprintf(":%s", env["PORT"])
	if err := http.ListenAndServeTLS(addr, env["SSLCertfilePath"], env["SSLKeyPath"], nil); err != nil {
		log.WithError(err).Fatal("無法啟動網頁伺服器")
	}
}

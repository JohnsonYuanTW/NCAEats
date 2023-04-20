package models

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/JohnsonYuanTW/NCAEats/config"
)

var db *gorm.DB

func init() {
	db = connect()
	initRestaurant()
	initMenuItem()
	initOrder()
	initOrderDetail()
}

func connect() *gorm.DB {
	env := config.Env
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Taipei",
		env["DB_URL"], env["DB_USERNAME"], env["DB_PASSWORD"], env["DB_NAME"], env["DB_PORT"])
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Error getting DB instance: %v", err)
		return nil
	}

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connected successfully!")
	return db
}

func GetDB() *gorm.DB {
	return db
}

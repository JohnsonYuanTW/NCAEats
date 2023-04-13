package models

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/JohnsonYuanTW/NCAEats/config"
)

var db *gorm.DB

func init() {
	connect()
	initRestaurant()
	initMenuItem()
	initOrder()
	initOrderDetail()
}

func connect() {
	env := config.Env

	dsn := "host=" + env["DB_URL"] + " user=" + env["DB_USERNAME"] + " password=" + env["DB_PASSWORD"] + " dbname=" + env["DB_NAME"] + " port=" + env["DB_PORT"] + " sslmode=disable TimeZone=Asia/Taipei"
	d, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}
	db = d
}

func GetDB() *gorm.DB {
	return db
}

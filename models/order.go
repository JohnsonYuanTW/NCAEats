package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Order struct {
	gorm.Model
	Owner        string
	RestaurantID int
	Restaurant   *Restaurant
}

func initOrder() {
	err := db.AutoMigrate(&Order{})
	if err != nil {
		panic("Order initialization failed. ")
	}
}

func (o *Order) CreateOrder() *Order {
	db.Create(&o)
	return o
}

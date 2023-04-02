package models

import (
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	owner        string `gorm:"" json:"owner"`
	RestaurantID int    `json:"restaruantID"`
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

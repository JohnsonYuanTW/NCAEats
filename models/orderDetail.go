package models

import (
	"gorm.io/gorm"
)

type OrderDetail struct {
	gorm.Model
	Owner      string
	OrderID    uint
	Order      *Order
	MenuItemID uint
	MenuItem   *MenuItem
}

func initOrderDetail() {
	err := db.AutoMigrate(&OrderDetail{})
	if err != nil {
		panic("OrderDetail initialization failed. ")
	}
}

func (od *OrderDetail) CreateOrderDetail() *OrderDetail {
	db.Create(&od)
	return od
}

func GetActiveOrderDetailsOfID(orderID uint) []OrderDetail {
	var orderDetails []OrderDetail
	db.Model(&OrderDetail{}).Where("order_id=?", orderID).Preload("MenuItem").Find(&orderDetails)
	return orderDetails
}

func DeleteOrderDetailsOfOrderID(orderID uint) []OrderDetail {
	var orderDetails []OrderDetail
	db.Model(&OrderDetail{}).Where("order_id=?", orderID).Delete(&orderDetails)
	return orderDetails
}

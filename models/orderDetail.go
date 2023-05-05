package models

import (
	"log"

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

func initOrderDetail(db *gorm.DB) (err error) {
	if err := db.AutoMigrate(&OrderDetail{}); err != nil {
		log.Fatalf("Failed to initialize OrderDetail: %v", err)
	}
	return
}

func (od *OrderDetail) CreateOrderDetail(db *gorm.DB) (*OrderDetail, error) {
	if err := db.Create(&od).Error; err != nil {
		return nil, err
	}
	return od, nil
}

func GetActiveOrderDetailsOfID(db *gorm.DB, orderID uint) ([]*OrderDetail, error) {
	var orderDetails []*OrderDetail
	result := db.Model(&OrderDetail{}).
		Where("order_id=?", orderID).
		Preload("MenuItem").
		Find(&orderDetails)
	return orderDetails, result.Error
}

func DeleteOrderDetailsOfOrderID(db *gorm.DB, orderID uint) error {
	var orderDetails []OrderDetail
	result := db.Model(&OrderDetail{}).Where("order_id=?", orderID).Delete(&orderDetails)
	return result.Error
}

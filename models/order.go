package models

import (
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Order struct {
	gorm.Model
	Owner        string
	RestaurantID uint
	Restaurant   *Restaurant
	OrderDetails []OrderDetail
}

func initOrder() (err error) {
	if err = db.AutoMigrate(&Order{}); err != nil {
		log.Fatalf("Error initializing Order: %v", err)
	}
	return
}

func (o *Order) CreateOrder() (*Order, error) {
	if err := db.Create(&o).Error; err != nil {
		return nil, err
	}
	return o, nil
}

func GetActiveOrders() ([]Order, error) {
	var orders []Order
	result := db.Model(&Order{}).Preload(clause.Associations).Find(&orders)
	return orders, result.Error
}

func GetActiveOrdersOfID(id string) ([]Order, error) {
	var orders []Order
	result := db.Model(&Order{}).Preload("Restaurant").Where("owner=?", id).Find(&orders)
	return orders, result.Error
}

func CountActiveOrderOfOwnerID(id string) (int64, error) {
	var count int64
	result := db.Model(&Order{}).Where("owner=?", id).Count(&count)
	return count, result.Error
}

func DeleteOrderOfID(id uint) error {
	result := db.Model(&Order{}).Where("id=?", id).Delete(&Order{})
	return result.Error
}

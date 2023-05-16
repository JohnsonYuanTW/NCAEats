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

func initOrder(db *gorm.DB) (err error) {
	if err = db.AutoMigrate(&Order{}); err != nil {
		log.Fatalf("Error initializing Order: %v", err)
	}
	return
}

func (o *Order) CreateOrder(db *gorm.DB) error {
	return db.Create(&o).Error
}

func GetActiveOrders(db *gorm.DB) ([]Order, error) {
	var orders []Order
	result := db.Model(&Order{}).Preload(clause.Associations).Find(&orders)
	return orders, result.Error
}

func GetActiveOrdersOfID(db *gorm.DB, id string) ([]Order, error) {
	var orders []Order
	result := db.Model(&Order{}).Preload("Restaurant").Where("owner=?", id).Find(&orders)
	return orders, result.Error
}

func CountActiveOrderOfOwnerID(db *gorm.DB, id string) (int64, error) {
	var count int64
	result := db.Model(&Order{}).Where("owner=?", id).Count(&count)
	return count, result.Error
}

func DeleteOrderOfID(db *gorm.DB, id uint) error {
	result := db.Model(&Order{}).Where("id=?", id).Delete(&Order{})
	return result.Error
}

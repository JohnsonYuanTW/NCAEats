package models

import (
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

func GetActiveOrders() []Order {
	var orders []Order
	db.Model(&Order{}).Preload(clause.Associations).Find(&orders)
	return orders
}

func GetActiveOrderOfID(id string) *Order {
	var order Order
	db.Model(&Order{}).Preload("Restaurant").Where("owner=?", id).Take(&order)
	return &order
}

// func CountActiveOrderOfID(id string) (int64, bool) {
// 	var ok bool
// 	var count int64
// 	if err := db.Model(&Order{}).Where("owner=?", id).Count(&count).Error; err != nil {
// 		// err
// 		ok = false
// 	} else {
// 		// no err
// 		ok = true
// 	}
// 	return count, ok
// }

func CountActiveOrderOfID(id string) (int64, bool) {
	var count int64
	err := db.Model(&Order{}).Where("owner=?", id).Count(&count).Error
	return count, err == nil
}

func DeleteOrderOfID(id string) *Order {
	var order Order
	if err := db.Model(&Order{}).Where("owner=?", id).Preload(clause.Associations).Delete(&order).Error; err != nil {
		return nil
	}
	return &order
}

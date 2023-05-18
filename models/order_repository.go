package models

import (
	"fmt"

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

type OrderRepository struct {
	*BaseRepository
}

func (r *OrderRepository) Init() (err error) {
	if err := r.DB.AutoMigrate(&Order{}); err != nil {
		return fmt.Errorf("error initializing Order: %v", err)
	}
	return nil
}

func (r *OrderRepository) CreateOrder(o *Order) error {
	return r.DB.Create(o).Error
}

func (r *OrderRepository) GetActiveOrders() ([]Order, error) {
	var orders []Order
	result := r.DB.Model(&Order{}).Preload(clause.Associations).Find(&orders)
	return orders, result.Error
}

func (r *OrderRepository) GetActiveOrdersOfID(id string) ([]Order, error) {
	var orders []Order
	result := r.DB.Model(&Order{}).Preload("Restaurant").Where("owner=?", id).Find(&orders)
	return orders, result.Error
}

func (r *OrderRepository) CountActiveOrderOfOwnerID(id string) (int64, error) {
	var count int64
	result := r.DB.Model(&Order{}).Where("owner=?", id).Count(&count)
	return count, result.Error
}

func (r *OrderRepository) DeleteOrderOfID(id uint) error {
	result := r.DB.Model(&Order{}).Where("id=?", id).Delete(&Order{})
	return result.Error
}

package models

import (
	"fmt"

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

type OrderDetailRepository struct {
	*BaseRepository
}

func (r *OrderDetailRepository) Init() (err error) {
	if err := r.DB.AutoMigrate(&OrderDetail{}); err != nil {
		return fmt.Errorf("error initializing OrderDetail: %v", err)
	}
	return
}

func (r *OrderDetailRepository) CreateOrderDetail(od *OrderDetail) error {
	return r.DB.Create(od).Error
}

func (r *OrderDetailRepository) GetActiveOrderDetailsOfID(orderID uint) ([]*OrderDetail, error) {
	var orderDetails []*OrderDetail
	result := r.DB.Model(&OrderDetail{}).
		Where("order_id=?", orderID).
		Preload("MenuItem").
		Find(&orderDetails)
	return orderDetails, result.Error
}

func (r *OrderDetailRepository) DeleteOrderDetailsOfOrderID(orderID uint) error {
	var orderDetails []OrderDetail
	result := r.DB.Model(&OrderDetail{}).Where("order_id=?", orderID).Delete(&orderDetails)
	return result.Error
}

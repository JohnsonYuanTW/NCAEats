package models

import (
	"fmt"

	"gorm.io/gorm"
)

// OrderDetail represents the details of a single order, including the menu items.
type OrderDetail struct {
	gorm.Model
	Owner      string
	OrderID    uint
	Order      *Order
	MenuItemID uint
	MenuItem   *MenuItem
}

// OrderDetailRepository defines the database operations for order details.
type OrderDetailRepository interface {
	Init() error
	CreateOrderDetail(*OrderDetail) error
	GetActiveOrderDetailsByOrderID(uint) ([]*OrderDetail, error)
	DeleteOrderDetailsByOrderID(uint) error
}

// OrderDetailGormRepository implements the OrderDetailRepository using the Gorm library.
type OrderDetailGormRepository struct {
	*BaseRepository
}

// Init initializes the order detail repository and performs auto-migrations.
func (r *OrderDetailGormRepository) Init() error {
	if err := r.DB.AutoMigrate(&OrderDetail{}); err != nil {
		return fmt.Errorf("failed to auto migrate OrderDetail: %w", err)
	}
	return nil
}

// CreateOrderDetail inserts a new order detail into the database.
func (r *OrderDetailGormRepository) CreateOrderDetail(od *OrderDetail) error {
	if err := r.DB.Create(od).Error; err != nil {
		return fmt.Errorf("failed to create order detail: %w", err)
	}
	return nil
}

// GetActiveOrderDetailsByOrderID fetches all active order details for a given order ID.
func (r *OrderDetailGormRepository) GetActiveOrderDetailsByOrderID(orderID uint) ([]*OrderDetail, error) {
	var orderDetails []*OrderDetail
	result := r.DB.
		Where("order_id=?", orderID).
		Preload("MenuItem").
		Find(&orderDetails)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch order details by order ID %d: %w", orderID, result.Error)
	}
	return orderDetails, nil
}

// DeleteOrderDetailsByOrderID removes all order details associated with a given order ID.
func (r *OrderDetailGormRepository) DeleteOrderDetailsByOrderID(orderID uint) error {
	var orderDetails []OrderDetail
	result := r.DB.Where("order_id=?", orderID).Delete(&orderDetails)
	if result.Error != nil {
		return fmt.Errorf("failed to delete order details by order ID %d: %w", orderID, result.Error)
	}
	return nil
}

package models

import (
	"fmt"
	"math/rand"

	"gorm.io/gorm"
)

// Order represents a restaurant order with associated details.
type Order struct {
	gorm.Model
	Owner        string
	ReportHTML   string
	ReportID     string
	RestaurantID uint
	Restaurant   *Restaurant
	OrderDetails []*OrderDetail
}

// OrderRepository provides an interface for database operations on orders.
type OrderRepository interface {
	Init() error
	CreateOrder(*Order) error
	GetActiveOrders() ([]*Order, error)
	GetActiveOrdersOfOwnerID(string) ([]*Order, error)
	CountActiveOrdersOfOwnerID(string) (int64, error)
	SaveOrderReport(uint, string) error
	GenerateUniqueReportID() string
	GetOrderReportByOrderID(uint) (string, error)
	GetOrderReportIDByOrderID(uint) (string, error)
	GetOrderIDByReportID(string) (uint, error)
	DeleteOrderByOrderID(uint) error
}

// OrderGormRepository implements the OrderRepository using the Gorm library.
type OrderGormRepository struct {
	*BaseRepository
}

// Init initializes the order repository and performs automigrations.
func (r *OrderGormRepository) Init() error {
	if err := r.DB.AutoMigrate(&Order{}); err != nil {
		return fmt.Errorf("failed to auto migrate Order: %w", err)
	}
	return nil
}

// CreateOrder inserts a new order into the database.
func (r *OrderGormRepository) CreateOrder(o *Order) error {
	if err := r.DB.Create(o).Error; err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

// GetActiveOrders fetches all active orders from the database.
func (r *OrderGormRepository) GetActiveOrders() ([]*Order, error) {
	var orders []*Order
	result := r.DB.Preload("Restaurant").Find(&orders)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch active orders: %w", result.Error)
	}
	return orders, nil
}

// GetActiveOrdersOfOwnerID fetches all active orders for a given owner ID.
func (r *OrderGormRepository) GetActiveOrdersOfOwnerID(ownerID string) ([]*Order, error) {
	var orders []*Order
	result := r.DB.Preload("Restaurant").Where("owner=?", ownerID).Find(&orders)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch orders for owner %s: %w", ownerID, result.Error)
	}
	return orders, nil
}

// CountActiveOrdersOfOwnerID counts all active orders for a given owner ID.
func (r *OrderGormRepository) CountActiveOrdersOfOwnerID(ownerID string) (int64, error) {
	var count int64
	result := r.DB.Model(&Order{}).Where("owner=?", ownerID).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count orders for owner %s: %w", ownerID, result.Error)
	}
	return count, nil
}

// SaveOrderReport updates an order with its report.
func (r *OrderGormRepository) SaveOrderReport(orderID uint, report string) error {
	order := &Order{}
	if err := r.DB.First(order, orderID).Error; err != nil {
		return err
	}

	reportID := r.GenerateUniqueReportID()

	order.ReportHTML = report
	order.ReportID = reportID

	if err := r.DB.Save(order).Error; err != nil {
		return fmt.Errorf("failed to save order report: %w", err)
	}
	return nil
}

// GenerateUniqueReportID generates a unique report ID, ensuring it does not exist in the database.
func (r *OrderGormRepository) GenerateUniqueReportID() string {
	var reportID string
	var count int64
	for {
		reportID = generateRandomID(6)
		r.DB.Model(&Order{}).Where("report_id = ?", reportID).Count(&count)
		if count == 0 {
			break
		}
	}
	return reportID
}

// GetOrderReportByOrderID retrieves the report HTML by order ID.
func (r *OrderGormRepository) GetOrderReportByOrderID(orderID uint) (string, error) {
	order := &Order{}
	if err := r.DB.First(&order, "id = ?", orderID).Error; err != nil {
		return "", err
	}
	return order.ReportHTML, nil
}

// GetOrderReportIDByOrderID retrieves the report ID by order ID.
func (r *OrderGormRepository) GetOrderReportIDByOrderID(orderID uint) (string, error) {
	order := &Order{}
	if err := r.DB.First(&order, "id = ?", orderID).Error; err != nil {
		return "", err
	}
	return order.ReportID, nil
}

// GetOrderIDByReportID retrieves the order ID by its report ID.
func (r *OrderGormRepository) GetOrderIDByReportID(reportID string) (uint, error) {
	order := &Order{}
	if err := r.DB.First(&order, "report_id = ?", reportID).Error; err != nil {
		return 0, err
	}
	return order.ID, nil
}

// DeleteOrderByOrderID deletes an order by its ID.
func (r *OrderGormRepository) DeleteOrderByOrderID(orderID uint) error {
	result := r.DB.Where("id=?", orderID).Delete(&Order{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete order with ID %d: %w", orderID, result.Error)
	}
	return nil
}

// generateRandomID is a helper function to generate random strings of a given length.
func generateRandomID(length int) string {
	const charset = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxy="

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

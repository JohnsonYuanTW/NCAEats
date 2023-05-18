package models

import (
	"fmt"
	"math/rand"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	Owner        string
	ReportHTML   string
	ReportID     string
	RestaurantID uint
	Restaurant   *Restaurant
	OrderDetails []OrderDetail
}

type OrderRepository struct {
	*BaseRepository
}

func (r *OrderRepository) Init() (err error) {
	if err := r.DB.AutoMigrate(&Order{}); err != nil {
		return fmt.Errorf("failed to auto migrate Order: %w", err)
	}
	return nil
}

func (r *OrderRepository) CreateOrder(o *Order) error {
	return r.DB.Create(o).Error
}

func (r *OrderRepository) GetActiveOrders() ([]Order, error) {
	var orders []Order
	result := r.DB.Preload("Restaurant").Find(&orders)
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

func (r *OrderRepository) SaveOrderReport(orderID uint, report string) error {
	order := &Order{}
	if err := r.DB.First(order, orderID).Error; err != nil {
		return err
	}

	reportID := r.generateUniqueReportID()

	order.ReportHTML = report
	order.ReportID = reportID

	if err := r.DB.Save(order).Error; err != nil {
		return err
	}
	return nil
}

func (r *OrderRepository) generateUniqueReportID() string {
	var reportID string
	var count int64
	for {
		reportID = generateRandomID(6)
		r.DB.Model(&Order{}).Where("report_id = ?", reportID).Count(&count)
		if count == 0 {
			return reportID
		}
	}
}

func (r *OrderRepository) GetOrderReportByOrderID(orderID uint, report string) (string, error) {
	order := &Order{}
	if err := r.DB.First(&order, "id = ?", orderID).Error; err != nil {
		return "", err
	}
	return order.ReportHTML, nil
}

func (r *OrderRepository) GetOrderReportIDByOrderID(orderID uint) (string, error) {
	order := &Order{}
	if err := r.DB.First(&order, "id = ?", orderID).Error; err != nil {
		return "", err
	}
	return order.ReportID, nil
}

func (r *OrderRepository) DeleteOrderByOrderID(orderID uint) error {
	result := r.DB.Where("id=?", orderID).Delete(&Order{})
	return result.Error
}

func generateRandomID(length int) string {
	const charset = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxy="

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

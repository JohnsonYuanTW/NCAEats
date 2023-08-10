package models

import (
	"fmt"

	"gorm.io/gorm"
)

// Restaurant represents a restaurant with its associated menu items and orders.
type Restaurant struct {
	gorm.Model
	Name      string
	Tel       string
	MenuItems []*MenuItem
	Orders    []*Order
}

// RestaurantRepository defines the database operations for restaurants.
type RestaurantRepository interface {
	Init() error
	CreateRestaurant(*Restaurant) error
	GetAllRestaurants() ([]*Restaurant, error)
	GetRestaurantByName(string) (*Restaurant, error)
	DeleteRestaurant(uint) error
}

// RestaurantGormRepository implements the RestaurantRepository using the Gorm library.
type RestaurantGormRepository struct {
	*BaseRepository
}

// Init initializes the restaurant repository and performs auto-migrations.
func (r *RestaurantGormRepository) Init() error {
	if err := r.DB.AutoMigrate(&Restaurant{}); err != nil {
		return fmt.Errorf("failed to auto migrate Restaurant: %w", err)
	}
	return nil
}

// CreateRestaurant inserts a new restaurant into the database.
func (r *RestaurantGormRepository) CreateRestaurant(rest *Restaurant) error {
	if err := r.DB.Create(rest).Error; err != nil {
		return fmt.Errorf("failed to create restaurant: %w", err)
	}
	return nil
}

// GetAllRestaurants fetches all restaurants from the database.
func (r *RestaurantGormRepository) GetAllRestaurants() ([]*Restaurant, error) {
	var restaurants []*Restaurant
	if err := r.DB.Find(&restaurants).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch all restaurants: %w", err)
	}
	return restaurants, nil
}

// GetRestaurantByName fetches a restaurant by its name from the database.
func (r *RestaurantGormRepository) GetRestaurantByName(name string) (*Restaurant, error) {
	var restaurant Restaurant
	if err := r.DB.Where("name=?", name).First(&restaurant).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch restaurant by name %s: %w", name, err)
	}
	return &restaurant, nil
}

// DeleteRestaurant removes a restaurant by its ID from the database.
func (r *RestaurantGormRepository) DeleteRestaurant(ID uint) error {
	var restaurant Restaurant
	result := r.DB.Where("ID=?", ID).Delete(&restaurant)
	if result.Error != nil {
		return fmt.Errorf("failed to delete restaurant with ID %d: %w", ID, result.Error)
	}
	return nil
}

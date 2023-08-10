package models

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MenuItem represents a single item on a restaurant's menu.
type MenuItem struct {
	gorm.Model
	Name         string
	Price        int
	RestaurantID uint
	Restaurant   *Restaurant
}

// MenuItemRepository defines the database operations for menu items.
type MenuItemRepository interface {
	Init() error
	CreateMenuItem(*MenuItem) error
	GetMenuItemsByRestaurantName(string) ([]*MenuItem, error)
	GetMenuItemByDetails(string, string) (*MenuItem, error)
}

// MenuItemGormRepository implements the MenuItemRepository using the Gorm library.
type MenuItemGormRepository struct {
	*BaseRepository
}

// Init initializes the menu item repository and performs auto-migrations.
func (r *MenuItemGormRepository) Init() error {
	if err := r.DB.AutoMigrate(&MenuItem{}); err != nil {
		return fmt.Errorf("failed to auto migrate MenuItem: %w", err)
	}
	return nil
}

// CreateMenuItem inserts a new menu item into the database.
func (r *MenuItemGormRepository) CreateMenuItem(mi *MenuItem) error {
	if err := r.DB.Create(mi).Error; err != nil {
		return fmt.Errorf("failed to create menu item: %w", err)
	}
	return nil
}

// GetMenuItemsByRestaurantName fetches all menu items for a given restaurant name.
func (r *MenuItemGormRepository) GetMenuItemsByRestaurantName(name string) ([]*MenuItem, error) {
	var restaurant Restaurant
	if err := r.DB.
		Preload(clause.Associations).
		Where("name = ?", name).
		Take(&restaurant).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch restaurant by name %s: %w", name, err)
	}
	return restaurant.MenuItems, nil
}

// GetMenuItemByDetails fetches a menu item based on its name and the name of its restaurant.
func (r *MenuItemGormRepository) GetMenuItemByDetails(itemName, restaurantName string) (*MenuItem, error) {
	var menuItem MenuItem
	err := r.DB.
		Preload("Restaurant", "name = ?", restaurantName).
		Where("name = ?", itemName).
		Take(&menuItem).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch menu item by details: %w", err)
	}

	if menuItem.Restaurant == nil || menuItem.Restaurant.Name != restaurantName {
		return nil, errors.New("no matching menu item found")
	}

	return &menuItem, nil
}

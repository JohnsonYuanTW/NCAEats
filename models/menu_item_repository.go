package models

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MenuItem struct {
	gorm.Model
	Name         string
	Price        int
	RestaurantID uint
	Restaurant   *Restaurant
}

type MenuItemRepository struct {
	*BaseRepository
}

func (r *MenuItemRepository) Init() error {
	if err := r.DB.AutoMigrate(&MenuItem{}); err != nil {
		return fmt.Errorf("error initializing MenuItem: %v", err)
	}
	return nil
}

func (r *MenuItemRepository) CreateMenuItem(mi *MenuItem) error {
	return r.DB.Create(mi).Error
}

func (r *MenuItemRepository) GetAllMenuItems() ([]MenuItem, error) {
	var menuItems []MenuItem
	result := r.DB.Model(&MenuItem{}).Find(&menuItems)
	return menuItems, result.Error
}

func (r *MenuItemRepository) GetMenuItemsByRestaurantName(name string) ([]MenuItem, error) {
	var restaurant Restaurant
	if err := r.DB.
		Preload(clause.Associations).
		Where("name = ?", name).
		Take(&restaurant).Error; err != nil {
		return nil, err
	}
	return restaurant.MenuItems, nil
}

func (r *MenuItemRepository) GetMenuItemByNameAndRestaurantName(itemName, restaurantName string) (*MenuItem, error) {
	var menuItem MenuItem
	err := r.DB.Model(&MenuItem{}).
		Preload("Restaurant", "name = ?", restaurantName).
		Where("name = ?", itemName).
		Take(&menuItem).Error

	if err != nil {
		return nil, err
	}

	return &menuItem, nil
}

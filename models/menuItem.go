package models

import (
	"log"

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

func initMenuItem(db *gorm.DB) (err error) {
	if err = db.AutoMigrate(&MenuItem{}); err != nil {
		log.Fatalf("Error initializing MenuItem: %v", err)
	}
	return
}

func (mi *MenuItem) CreateMenuItem(db *gorm.DB) error {
	return db.Create(&mi).Error
}

func GetAllMenuItems(db *gorm.DB) ([]MenuItem, error) {
	var menuItems []MenuItem
	result := db.Model(&MenuItem{}).Find(&menuItems)
	return menuItems, result.Error
}

func GetMenuItemsByRestaurantName(db *gorm.DB, name string) ([]MenuItem, error) {
	var restaurant Restaurant
	if err := db.
		Preload(clause.Associations).
		Where("name = ?", name).
		Take(&restaurant).Error; err != nil {
		return nil, err
	}
	return restaurant.MenuItems, nil
}

func GetMenuItemByNameAndRestaurantName(db *gorm.DB, itemName, restaurantName string) (*MenuItem, error) {
	var menuItem MenuItem
	err := db.Model(&MenuItem{}).
		Preload("Restaurant", "name = ?", restaurantName).
		Where("name = ?", itemName).
		Take(&menuItem).Error

	if err != nil {
		return nil, err
	}

	return &menuItem, nil
}

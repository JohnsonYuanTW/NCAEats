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

func initMenuItem() (err error) {
	if err = db.AutoMigrate(&MenuItem{}); err != nil {
		log.Fatalf("Error initializing MenuItem: %v", err)
	}
	return
}

func (mi *MenuItem) CreateMenuItem() (*MenuItem, error) {
	if err := db.Create(&mi).Error; err != nil {
		return nil, err
	}
	return mi, nil
}

func GetAllMenuItems() ([]MenuItem, error) {
	var menuItems []MenuItem
	result := db.Model(&MenuItem{}).Find(&menuItems)
	return menuItems, result.Error
}

func GetMenuItemsByRestaurantName(name string) ([]MenuItem, error) {
	var restaurant Restaurant
	if err := db.
		Preload(clause.Associations).
		Where("name = ?", name).
		Take(&restaurant).Error; err != nil {
		return nil, err
	}
	return restaurant.MenuItems, nil
}

func GetMenuItemByNameAndRestaurantName(itemName, restaurantName string) (*MenuItem, error) {
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

package models

import (
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

func initMenuItem() {
	err := db.AutoMigrate(&MenuItem{})
	if err != nil {
		panic("MenuItem initialization failed. ")
	}
}

func (mi *MenuItem) CreateMenuItem() *MenuItem {
	db.Create(&mi)
	return mi
}

func GetAllMenuItems() []MenuItem {
	var menuItems []MenuItem
	db.Model(&MenuItem{}).Find(&menuItems)
	return menuItems
}

func GetMenuItemsByRestaurantName(name string) []MenuItem {
	restaurant := &Restaurant{}
	if err := db.Model(&Restaurant{}).Preload(clause.Associations).Where("name=?", name).Take(&restaurant).Error; err != nil {
		return nil
	}
	return restaurant.MenuItems
}

func GetMenuItemByNameAndRestaurantName(itemName string, restaurantName string) (*MenuItem, bool) {
	var menuItem MenuItem
	ok := false
	if err := db.Model(&MenuItem{}).Preload("Restaurant", "name=?", restaurantName).Where("name=?", itemName).Take(&menuItem).Error; err == nil {
		ok = true
	}
	return &menuItem, ok
}

func DeleteMenuItem(ID int64) *MenuItem {
	var menuItem MenuItem
	db.Model(&MenuItem{}).Where("ID=?", ID).Delete(&menuItem)
	return &menuItem
}

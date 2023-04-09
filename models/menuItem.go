package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MenuItem struct {
	gorm.Model
	Name         string
	Price        int
	RestaurantID int
	Restaurant   *Restaurant
}

func initMenuItem() {
	err := db.AutoMigrate(&MenuItem{})
	if err != nil {
		panic("MenuItem initialization failed. ")
	}
}

func (r *MenuItem) CreateMenuItem() *MenuItem {
	db.Create(&r)
	return r
}

func GetAllMenuItems() []MenuItem {
	var menuItems []MenuItem
	db.Model(&MenuItem{}).Find(&menuItems)
	return menuItems
}

func GetMenuItemsByRestaurantID(restaurantID uint) []MenuItem {
	restaurant := &Restaurant{}
	if err := db.Model(&Restaurant{}).Preload(clause.Associations).Take(&restaurant, restaurantID).Error; err != nil {
		return nil
	}
	return restaurant.MenuItems
}

func GetMenuItemByName(name string) (*MenuItem, bool) {
	var menuItem MenuItem
	ok := false
	if err := db.Model(&MenuItem{}).Where("name=?", name).First(&menuItem).Error; err == nil {
		ok = true
	}
	return &menuItem, ok
}

func DeleteMenuItem(ID int64) *MenuItem {
	var menuItem MenuItem
	db.Model(&MenuItem{}).Where("ID=?", ID).Delete(&menuItem)
	return &menuItem
}

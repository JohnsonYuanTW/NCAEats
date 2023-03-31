package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MenuItem struct {
	gorm.Model
	Name         string      `gorm:"" json:"name"`
	Price        int         `json:"price"`
	RestaurantID int         `json:"restaruantID"`
	Restaurant   *Restaurant `json:"restaurant,omitempty"`
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
	var MenuItems []MenuItem
	db.Find(&MenuItems)
	return MenuItems
}

func GetMenuItemsByRestaurantID(restaurantID uint) ([]MenuItem, *gorm.DB) {
	restaurant := &Restaurant{}
	if err := db.Preload(clause.Associations).Take(&restaurant, restaurantID).Error; err != nil {
		return nil, db
	}
	return restaurant.MenuItems, db
}

func GetMenuItemByName(name string) (*MenuItem, *gorm.DB, bool) {
	var menuItem MenuItem
	ok := false
	if err := db.Where("name=?", name).First(&menuItem).Error; err == nil {
		ok = true
	}
	return &menuItem, db, ok
}

func DeleteMenuItem(ID int64) *MenuItem {
	var menuItem MenuItem
	db.Where("ID=?", ID).Delete(&menuItem)
	return &menuItem
}

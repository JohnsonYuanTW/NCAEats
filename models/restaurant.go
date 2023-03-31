package models

import (
	"gorm.io/gorm"
)

type Restaurant struct {
	gorm.Model
	Name      string     `gorm:"" json:"name"`
	Tel       string     `json:"tel"`
	MenuItems []MenuItem `json:"menuitems"`
}

func initRestaurant() {
	err := db.AutoMigrate(&Restaurant{})
	if err != nil {
		panic("Restaurant initialization failed. ")
	}
}

func (r *Restaurant) CreateRestaurant() *Restaurant {
	db.Create(&r)
	return r
}

func GetAllRestaurants() []Restaurant {
	var Restaurants []Restaurant
	db.Find(&Restaurants)
	return Restaurants
}

func GetRestaurantByName(name string) (*Restaurant, *gorm.DB, bool) {
	var restaurant Restaurant
	ok := true
	if err := db.Where("name=?", name).First(&restaurant).Error; err != nil {
		ok = false
	}
	return &restaurant, db, ok
}

func DeleteRestaurant(ID int64) *Restaurant {
	var restaurant Restaurant
	db.Where("ID=?", ID).Delete(&restaurant)
	return &restaurant
}

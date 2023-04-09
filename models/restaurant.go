package models

import (
	"gorm.io/gorm"
)

type Restaurant struct {
	gorm.Model
	Name      string
	Tel       string
	MenuItems []MenuItem
	Orders    []Order
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
	var restaurants []Restaurant
	db.Model(&Restaurant{}).Find(&restaurants)
	return restaurants
}

func GetRestaurantByName(name string) (*Restaurant, bool) {
	var restaurant Restaurant
	ok := true
	if err := db.Model(&Restaurant{}).Where("name=?", name).First(&restaurant).Error; err != nil {
		ok = false
	}
	return &restaurant, ok
}

func DeleteRestaurant(ID int64) *Restaurant {
	var restaurant Restaurant
	db.Model(&Restaurant{}).Where("ID=?", ID).Delete(&restaurant)
	return &restaurant
}

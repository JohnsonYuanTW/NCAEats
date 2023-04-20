package models

import (
	"log"

	"gorm.io/gorm"
)

type Restaurant struct {
	gorm.Model
	Name      string
	Tel       string
	MenuItems []MenuItem
	Orders    []Order
}

func initRestaurant() (err error) {
	if err = db.AutoMigrate(&Restaurant{}); err != nil {
		log.Fatalf("Error initializing Restaurant: %v", err)
	}
	return
}

func (r *Restaurant) CreateRestaurant() (*Restaurant, error) {
	if err := db.Create(&r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func GetAllRestaurants() ([]Restaurant, error) {
	var restaurants []Restaurant
	if err := db.Model(&Restaurant{}).Find(&restaurants).Error; err != nil {
		return nil, err
	}
	return restaurants, nil
}

func GetRestaurantByName(name string) (*Restaurant, error) {
	var restaurant Restaurant
	if err := db.Model(&Restaurant{}).Where("name=?", name).First(&restaurant).Error; err != nil {
		return nil, err
	}
	return &restaurant, nil
}

func DeleteRestaurant(ID int64) error {
	var restaurant Restaurant
	result := db.Model(&Restaurant{}).Where("ID=?", ID).Delete(&restaurant)
	return result.Error
}

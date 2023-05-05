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

func initRestaurant(db *gorm.DB) (err error) {
	if err = db.AutoMigrate(&Restaurant{}); err != nil {
		log.Fatalf("Error initializing Restaurant: %v", err)
	}
	return
}

func (r *Restaurant) CreateRestaurant(db *gorm.DB) error {
	return db.Create(&r).Error
}

func GetAllRestaurants(db *gorm.DB) ([]*Restaurant, error) {
	var restaurants []*Restaurant
	if err := db.Find(&restaurants).Error; err != nil {
		return nil, err
	}
	return restaurants, nil
}

func GetRestaurantByName(db *gorm.DB, name string) (*Restaurant, error) {
	var restaurant Restaurant
	if err := db.Model(&Restaurant{}).Where("name=?", name).First(&restaurant).Error; err != nil {
		return nil, err
	}
	return &restaurant, nil
}

func DeleteRestaurant(db *gorm.DB, ID int64) error {
	var restaurant Restaurant
	result := db.Model(&Restaurant{}).Where("ID=?", ID).Delete(&restaurant)
	return result.Error
}

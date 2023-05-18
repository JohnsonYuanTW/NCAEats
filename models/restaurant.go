package models

import (
	"fmt"

	"gorm.io/gorm"
)

type Restaurant struct {
	gorm.Model
	Name      string
	Tel       string
	MenuItems []MenuItem
	Orders    []Order
}

type RestaurantRepository struct {
	*BaseRepository
}

func (r *RestaurantRepository) Init() error {
	if err := r.DB.AutoMigrate(&Restaurant{}); err != nil {
		return fmt.Errorf("error initializing Restaurant: %v", err)
	}
	return nil
}

func (r *RestaurantRepository) CreateRestaurant(rest *Restaurant) error {
	return r.DB.Create(rest).Error
}

func (r *RestaurantRepository) GetAllRestaurants() ([]*Restaurant, error) {
	var restaurants []*Restaurant
	if err := r.DB.Find(&restaurants).Error; err != nil {
		return nil, err
	}
	return restaurants, nil
}

func (r *RestaurantRepository) GetRestaurantByName(name string) (*Restaurant, error) {
	var restaurant Restaurant
	if err := r.DB.Model(&Restaurant{}).Where("name=?", name).First(&restaurant).Error; err != nil {
		return nil, err
	}
	return &restaurant, nil
}

func (r *RestaurantRepository) DeleteRestaurant(ID int64) error {
	var restaurant Restaurant
	result := r.DB.Model(&Restaurant{}).Where("ID=?", ID).Delete(&restaurant)
	return result.Error
}

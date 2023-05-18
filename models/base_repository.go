package models

import "gorm.io/gorm"

type BaseRepository struct {
	DB *gorm.DB
}

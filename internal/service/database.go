package service

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	DB = db
	return db, nil
}

func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		log.Printf("DB not initialized")
		return nil
	}
	return DB.AutoMigrate(models...)
}

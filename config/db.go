package config

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vaidikcode/minipay/models"
)

var DB *gorm.DB

func InitDB(path string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Transaction{}, &models.WebhookEvent{}, &models.IdempotencyKey{}); err != nil {
		log.Fatal(err)
	}
	DB = db
	return db
}

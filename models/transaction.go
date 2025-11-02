package models

import "time"

type Transaction struct {
	ID        string    `gorm:"primaryKey"`
	Amount    int64     `gorm:"not null"`
	Currency  string    `gorm:"size:8;not null;default:'usd'"`
	Customer  string    `gorm:"size:64;index"`
	Status    string    `gorm:"size:32;index;default:'pending'"`
	Refunded  bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (t Transaction) TableName() string {
	return "transactions"
}

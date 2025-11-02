package models

import "time"

type IdempotencyKey struct {
	ID            string    `gorm:"primaryKey"`
	TransactionID string    `gorm:"index;not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}

func (i IdempotencyKey) TableName() string {
	return "idempotency_keys"
}

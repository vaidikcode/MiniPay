package models

import "time"

type WebhookEvent struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	TransactionID string    `gorm:"index;not null"`
	EventType     string    `gorm:"size:64;not null"`
	Payload       string    `gorm:"type:text;not null"`
	TargetURL     string    `gorm:"size:512;not null"`
	Status        string    `gorm:"size:32;index;default:'pending'"`
	Attempts      int       `gorm:"default:0"`
	NextRunAt     time.Time `gorm:"index"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (w WebhookEvent) TableName() string {
	return "webhook_events"
}

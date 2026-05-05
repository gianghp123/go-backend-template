package models

import "time"

type Example struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"type:varchar(255)"`
	Description string    `gorm:"type:text"`
	OwnerID     string    `gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Example) TableName() string { return "examples" }

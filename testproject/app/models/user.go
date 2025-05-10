package models

import (
	"time"
)

// User represents a user entity
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	// Add your custom fields here
	Name string `json:"name" gorm:"size:255;not null"`
	// Add more fields as needed
}

// TableName overrides the default table name
func (User) TableName() string {
	return "users"
}

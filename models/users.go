package models

import "time"

type Users struct {
	ID        uint      `json:"id" gorm:"primaryKey:autoIncrement"`
	Username  string    `json:"username" gorm:"type:varchar(255)"`
	FullName  string    `json:"full_name" gorm:"type:varchar(255)"`
	Password  string    `json:"-" gorm:"type:varchar(255)"`
	Mail      string    `json:"mail" gorm:"type:varchar(255);uniqueIndex"`
	Role      string    `json:"role" gorm:"type:varchar(55)"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

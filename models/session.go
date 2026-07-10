package models

import "time"

type Session struct {
	ID          uint       `json:"id" gorm:"primaryKey:autoIncrement"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	Description string     `json:"description" gorm:"type:text"`
	Type        string     `json:"type" gorm:"type:varchar(50);not null"`
	Category    string     `json:"category" gorm:"type:varchar(50);not null"`
	Points      int        `json:"points" gorm:"default:10"`
	QRCodeKey   string     `json:"qr_code_key" gorm:"type:varchar(255);uniqueIndex;not null"`
	Tags        []string   `json:"tags" gorm:"serializer:json"`
	RoomName    string     `json:"room_name" gorm:"type:varchar(100)"`
	SpeakerName string     `json:"speaker_name" gorm:"type:varchar(100)"`
	SpeakerTitle string    `json:"speaker_title" gorm:"type:varchar(100)"`
	StartTime   *time.Time `json:"start_time" gorm:"type:timestamp"`
	EndTime     *time.Time `json:"end_time" gorm:"type:timestamp"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type Scan struct {
	ID        uint      `json:"id" gorm:"primaryKey:autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_user_session"`
	SessionID uint      `json:"session_id" gorm:"not null;uniqueIndex:idx_user_session"`
	ScannedAt time.Time `json:"scanned_at" gorm:"autoCreateTime"`

	User    Users   `json:"-" gorm:"foreignKey:UserID"`
	Session Session `json:"session" gorm:"foreignKey:SessionID"`
}

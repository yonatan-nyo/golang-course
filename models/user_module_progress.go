package models

import "time"

type UserModuleProgress struct {
	ID          string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      string     `json:"user_id" gorm:"not null"`
	ModuleID    string     `json:"module_id" gorm:"not null"`
	IsCompleted bool       `json:"is_completed" gorm:"default:false"`
	CompletedAt *time.Time `json:"completed_at"`

	User   User   `json:"-" gorm:"foreignKey:UserID"`
	Module Module `json:"-" gorm:"foreignKey:ModuleID"`
}

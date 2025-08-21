package models

import (
	"time"

	"gorm.io/gorm"
)

type UserCourse struct {
	ID          string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      string    `json:"user_id" gorm:"not null"`
	CourseID    string    `json:"course_id" gorm:"not null"`
	PurchasedAt time.Time `json:"purchased_at"`

	User   User   `json:"-" gorm:"foreignKey:UserID"`
	Course Course `json:"-" gorm:"foreignKey:CourseID"`
}

func (uc *UserCourse) BeforeCreate(tx *gorm.DB) error {
	uc.PurchasedAt = time.Now()
	return nil
}

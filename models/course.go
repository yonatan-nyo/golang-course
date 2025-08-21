package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Course struct {
	ID          string         `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description"`
	Instructor  string         `json:"instructor" gorm:"not null"`
	Price       float64        `json:"price" gorm:"not null"`
	Thumbnail   string         `json:"thumbnail"`
	Topics      pq.StringArray `json:"topics" gorm:"type:text[]"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Modules []Module `json:"modules" gorm:"foreignKey:CourseID"`
}

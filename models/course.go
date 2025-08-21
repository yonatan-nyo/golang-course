package models

import "time"

type Course struct {
	ID             string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title          string    `json:"title" gorm:"not null"`
	Description    string    `json:"description" gorm:"not null"`
	Instructor     string    `json:"instructor" gorm:"not null"`
	Topics         []string  `json:"topics" gorm:"type:text[]"`
	Price          float64   `json:"price" gorm:"not null"`
	ThumbnailImage *string   `json:"thumbnail_image"`
	TotalModules   int       `json:"total_modules" gorm:"default:0"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	Modules []Module `json:"modules,omitempty" gorm:"foreignKey:CourseID"`
}

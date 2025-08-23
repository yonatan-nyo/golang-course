package models

import "time"

type Module struct {
	ID           string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CourseID     string    `json:"course_id" gorm:"not null"`
	Title        string    `json:"title" gorm:"not null"`
	Description  string    `json:"description" gorm:"not null"`
	Order        int       `json:"order" gorm:"not null"`
	PDFContent   *string   `json:"pdf_content"`
	VideoContent *string   `json:"video_content"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Course Course `json:"-" gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
}

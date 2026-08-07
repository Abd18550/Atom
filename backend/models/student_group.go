package models

import (
	"time"
)

type StudentGroup struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SchoolName   string    `gorm:"not null" json:"school_name"`
	Class        string    `gorm:"not null" json:"class"`
	AcademicYear string    `gorm:"not null" json:"academic_year"`
	CreatedByID  uint      `gorm:"not null;default:1" json:"created_by_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

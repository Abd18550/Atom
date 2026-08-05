package models

import "time"

// Exercise represents a coding challenge for students
type Exercise struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Title         string    `json:"title" gorm:"not null"`
	Description   string    `json:"description" gorm:"type:text;not null"`
	Type          string    `json:"type" gorm:"default:'program'"` // "program" or "function"
	RightSolution string    `json:"right_solution" gorm:"type:text;not null"`
	Test          string    `json:"test" gorm:"type:text;not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

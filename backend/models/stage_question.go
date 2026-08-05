package models

import "time"

// StageQuestion represents a coding challenge explicitly tied to a learning path stage.
type StageQuestion struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Title         string    `json:"title" gorm:"not null"`
	Description   string    `json:"description" gorm:"type:text;not null"`
	Type          string    `json:"type" gorm:"default:'program'"`
	RightSolution string    `json:"right_solution" gorm:"type:text;not null"`
	Test          string    `json:"test" gorm:"type:text;not null"`
	StageID       uint      `json:"stage_id" gorm:"not null"`
	OrderIndex    int       `json:"order_index" gorm:"default:0"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

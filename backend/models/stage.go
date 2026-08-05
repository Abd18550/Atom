package models

import "time"

// Stage represents a learning stage which groups multiple exercises sequentially
type Stage struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description" gorm:"type:text"`
	OrderIndex  int        `json:"order_index" gorm:"not null;unique"` // 1 for Stage 1, 2 for Stage 2, etc.
	Questions   []StageQuestion `json:"questions" gorm:"foreignKey:StageID"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

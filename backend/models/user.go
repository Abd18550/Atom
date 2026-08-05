package models

import (
	"time"
)

type User struct {
	ID             uint          `gorm:"primaryKey" json:"id"`
	Username       string        `gorm:"uniqueIndex;not null" json:"username"`
	Email          string        `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash   string        `gorm:"not null" json:"-"`
	Role           string        `gorm:"not null" json:"role"` // "Admin", "Mentor", "Student"
	Avatar         string        `json:"avatar"`
	FullName       string        `gorm:"not null" json:"full_name"`
	StudentGroupID *uint         `json:"student_group_id,omitempty"`
	StudentGroup   *StudentGroup `json:"student_group,omitempty" gorm:"foreignKey:StudentGroupID"`
	DateOfBirth    *string       `json:"date_of_birth,omitempty"`
	Theme          string        `json:"theme"`
	XP             int           `json:"xp" gorm:"default:0"`
	Level          int           `json:"level" gorm:"default:1"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

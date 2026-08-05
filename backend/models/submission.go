package models

import "time"

// Submission tracks a student's attempt to solve an exercise
type Submission struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"not null"`
	// Regular exercise is nullable if it's a stage question submission
	ExerciseID     *uint     `json:"exercise_id"` 
	StageQuestionID *uint    `json:"stage_question_id"`
	Code           string    `json:"code" gorm:"type:text;not null"`
	Status         string    `json:"status" gorm:"default:'Pending'"` // Pending, Running, Passed, Failed, SyntaxError
	ErrorMessage   string    `json:"error_message" gorm:"type:text"`
	ExpectedOutput string    `json:"expected_output" gorm:"type:text"`
	ActualOutput   string    `json:"actual_output" gorm:"type:text"`
	FailedTestcase string    `json:"failed_testcase" gorm:"type:text"` // Which input failed
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

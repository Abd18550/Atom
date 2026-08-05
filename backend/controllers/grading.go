package controllers

import (
	//"fmt"
	"net/http"
	//"os"

	"backend/database"
	"backend/models"
	"backend/worker"

	"github.com/gin-gonic/gin"
)

// GetExercises returns a list of all coding exercises
func GetExercises(c *gin.Context) {
	var exercises []models.Exercise
	// Do not send right solutions or testcases to the frontend list for security and size
	if err := database.DB.Select("id", "title", "description", "created_at").Find(&exercises).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exercises"})
		return
	}
	c.JSON(http.StatusOK, exercises)
}

// GetExerciseByID returns a specific exercise. Admins get all fields, students get restricted fields.
func GetExerciseByID(c *gin.Context) {
	id := c.Param("id")
	var exercise models.Exercise

	role, _ := c.Get("role")

	if role == "Admin" || role == "Supervisor" {
		// Admin gets all fields for editing
		if err := database.DB.First(&exercise, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
			return
		}
	} else {
		// Students only get fields needed for solving
		if err := database.DB.Select("id", "title", "description", "type", "created_at").First(&exercise, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
			return
		}
	}

	c.JSON(http.StatusOK, exercise)
}

type SubmitCodeInput struct {
	ExerciseID      *uint  `json:"exercise_id"`
	StageQuestionID *uint  `json:"stage_question_id"`
	Code            string `json:"code" binding:"required"`
}

// SubmitCode creates a new submission and adds it to the grading queue
func SubmitCode(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input SubmitCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.ExerciseID == nil && input.StageQuestionID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either exercise_id or stage_question_id must be provided"})
		return
	}

	// Create Submission
	submission := models.Submission{
		UserID:          userID.(uint),
		ExerciseID:      input.ExerciseID,
		StageQuestionID: input.StageQuestionID,
		Code:            input.Code,
		Status:          "Pending",
	}

	if err := database.DB.Create(&submission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save submission"})
		return
	}

	// Send to background worker queue
	worker.SubmissionQueue <- submission.ID

	c.JSON(http.StatusOK, gin.H{
		"message":       "Submission received and queued for grading.",
		"submission_id": submission.ID,
	})
}

// GetSubmissionStatus returns the current status and output of a submission
func GetSubmissionStatus(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	var submission models.Submission
	if err := database.DB.First(&submission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	// Security: Only the user who submitted it (or an Admin) should see it.
	// For now, restrict to submitter.
	if submission.UserID != userID.(uint) {
		role, _ := c.Get("role")
		if role != "Admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to view this submission."})
			return
		}
	}

	c.JSON(http.StatusOK, submission)
}

type CreateExerciseInput struct {
	Title         string `json:"title" binding:"required"`
	Description   string `json:"description" binding:"required"`
	Type          string `json:"type" binding:"required"` // 'program' or 'function'
	RightSolution string `json:"right_solution" binding:"required"`
	Test          string `json:"test" binding:"required"` // Raw strings from frontend
}

// CreateExercise allows Admins to create new challenges
func CreateExercise(c *gin.Context) {
	// 1. Authorization: Only allow Admin or Supervisor
	role, exists := c.Get("role")
	if !exists || (role != "Admin" && role != "Supervisor") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized. Only Admins and Supervisors can create challenges."})
		return
	}

	// 2. Parse Input
	var input CreateExerciseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format."})
		return
	}

	// 3. Create Exercise Model
	newEx := models.Exercise{
		Title:         input.Title,
		Description:   input.Description,
		Type:          input.Type,
		RightSolution: input.RightSolution,
		Test:          input.Test,
	}

	if err := database.DB.Create(&newEx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exercise record in database."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Exercise created securely.",
		"exercise": newEx,
	})
}

// UpdateExercise allows Admins to edit existing challenges
func UpdateExercise(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || (role != "Admin" && role != "Supervisor") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized. Only Admins and Supervisors can edit challenges."})
		return
	}

	id := c.Param("id")
	var exercise models.Exercise
	if err := database.DB.First(&exercise, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found."})
		return
	}

	var input CreateExerciseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format."})
		return
	}

	exercise.Title = input.Title
	exercise.Description = input.Description
	exercise.Type = input.Type
	exercise.RightSolution = input.RightSolution
	exercise.Test = input.Test

	if err := database.DB.Save(&exercise).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exercise in database."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Exercise updated successfully.",
		"exercise": exercise,
	})
}

// DeleteExercise allows Admins to remove challenges
func DeleteExercise(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || (role != "Admin" && role != "Supervisor") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized. Only Admins and Supervisors can delete challenges."})
		return
	}

	id := c.Param("id")
	var exercise models.Exercise
	if err := database.DB.First(&exercise, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found."})
		return
	}

	// GORM will soft or hard delete based on model config. We'll manually delete related submissions first to be safe.
	database.DB.Where("exercise_id = ?", id).Delete(&models.Submission{})

	if err := database.DB.Delete(&exercise).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete exercise."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exercise and related submissions deleted successfully."})
}

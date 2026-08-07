package controllers

import (
	"backend/database"
	"backend/models"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

// LevelInfo contains the title and XP threshold for a level
type LevelInfo struct {
	Level     int    `json:"level"`
	Title     string `json:"title"`
	Threshold int    `json:"threshold"`
}

// GetLevelTable returns the level titles and thresholds
var levelTable = []LevelInfo{
	{1, "Beginner", 0},
	{2, "Learner", 100},
	{3, "Explorer", 250},
	{4, "Coder", 500},
	{5, "Developer", 800},
	{6, "Engineer", 1200},
	{7, "Architect", 1800},
	{8, "Master", 2500},
	{9, "Grandmaster", 3500},
	{10, "Legend", 5000},
}

func getLevelTitle(level int) string {
	if level < 1 || level > len(levelTable) {
		return "Unknown"
	}
	return levelTable[level-1].Title
}

func getNextLevelXP(level int) int {
	if level >= len(levelTable) {
		return levelTable[len(levelTable)-1].Threshold // Max level
	}
	return levelTable[level].Threshold
}

func getCurrentLevelXP(level int) int {
	if level < 1 || level > len(levelTable) {
		return 0
	}
	return levelTable[level-1].Threshold
}

// GetStudentStats returns detailed progression stats for the student home page
func GetStudentStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Fetch user data
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Count total stages
	var totalStages int64
	database.DB.Model(&models.Stage{}).Count(&totalStages)

	// Fetch all stages with questions to determine completion
	var stages []models.Stage
	database.DB.Preload("Questions").Find(&stages)
	sort.Slice(stages, func(i, j int) bool {
		return stages[i].OrderIndex < stages[j].OrderIndex
	})

	// Get passed stage question IDs
	var passedSubmissions []models.Submission
	database.DB.Select("stage_question_id").
		Where("user_id = ? AND status = ? AND stage_question_id IS NOT NULL", userID, "Passed").
		Group("stage_question_id").
		Find(&passedSubmissions)

	passedMap := make(map[uint]bool)
	for _, sub := range passedSubmissions {
		if sub.StageQuestionID != nil {
			passedMap[*sub.StageQuestionID] = true
		}
	}

	// Calculate completed stages and find current stage
	completedStages := 0
	var currentStageID uint
	var currentStageTitle string
	totalQuestions := 0
	passedQuestions := len(passedMap)

	for _, stage := range stages {
		totalQuestions += len(stage.Questions)
		allPassed := len(stage.Questions) > 0
		for _, q := range stage.Questions {
			if !passedMap[q.ID] {
				allPassed = false
			}
		}
		if allPassed {
			completedStages++
		} else if currentStageID == 0 {
			currentStageID = stage.ID
			currentStageTitle = stage.Title
		}
	}

	// Count passed exercises (independent coding challenges)
	var passedExercises int64
	database.DB.Model(&models.Submission{}).
		Where("user_id = ? AND status = ? AND exercise_id IS NOT NULL", userID, "Passed").
		Group("exercise_id").
		Count(&passedExercises)

	// Level info
	levelTitle := getLevelTitle(user.Level)
	currentLevelXP := getCurrentLevelXP(user.Level)
	nextLevelXP := getNextLevelXP(user.Level)

	c.JSON(http.StatusOK, gin.H{
		"xp":                 user.XP,
		"level":              user.Level,
		"level_title":        levelTitle,
		"current_level_xp":   currentLevelXP,
		"next_level_xp":      nextLevelXP,
		"total_stages":       totalStages,
		"completed_stages":   completedStages,
		"current_stage_id":   currentStageID,
		"current_stage_title": currentStageTitle,
		"total_questions":    totalQuestions,
		"passed_questions":   passedQuestions,
		"passed_exercises":   passedExercises,
	})
}

package controllers

import (
	"backend/database"
	"backend/models"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

type ExercisePathResponse struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	OrderIndex  int    `json:"order_index"`
	Passed      bool   `json:"passed"`
}

type StagePathResponse struct {
	ID          uint                   `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	OrderIndex  int                    `json:"order_index"`
	Unlocked    bool                   `json:"unlocked"`
	Exercises   []ExercisePathResponse `json:"exercises"`
}

// GetLearningPath calculates student's progression and returns the pathway
func GetLearningPath(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var stages []models.Stage
	database.DB.Preload("Questions").Find(&stages)

	// Sort stages by OrderIndex
	sort.Slice(stages, func(i, j int) bool {
		return stages[i].OrderIndex < stages[j].OrderIndex
	})

	var passedSubmissions []models.Submission
	// Retrieve unique stage questions the user has passed
	database.DB.Select("stage_question_id").Where("user_id = ? AND status = ? AND stage_question_id IS NOT NULL", userID, "Passed").Group("stage_question_id").Find(&passedSubmissions)

	passedMap := make(map[uint]bool)
	for _, sub := range passedSubmissions {
		if sub.StageQuestionID != nil {
			passedMap[*sub.StageQuestionID] = true
		}
	}
	// Determine if we should bypass locks for Admin/Supervisor/Mentor
	role, _ := c.Get("role")
	isBypassRole := role != "Student"

	var response []StagePathResponse
	unlockNext := true // First stage is accessible initially

	for _, stage := range stages {
		stageResp := StagePathResponse{
			ID:          stage.ID,
			Title:       stage.Title,
			Description: stage.Description,
			OrderIndex:  stage.OrderIndex,
			Unlocked:    unlockNext || isBypassRole,
			Exercises:   make([]ExercisePathResponse, 0),
		}

		sort.Slice(stage.Questions, func(i, j int) bool {
			return stage.Questions[i].OrderIndex < stage.Questions[j].OrderIndex
		})

		allQuestionsPassed := true
		for _, q := range stage.Questions {
			passed := passedMap[q.ID] || isBypassRole
			if !passedMap[q.ID] {
				allQuestionsPassed = false
			}
			stageResp.Exercises = append(stageResp.Exercises, ExercisePathResponse{
				ID:          q.ID,
				Title:       q.Title,
				Description: q.Description,
				OrderIndex:  q.OrderIndex,
				Passed:      passed,
			})
		}

		if len(stage.Questions) == 0 {
			allQuestionsPassed = false
		}

		response = append(response, stageResp)

		// Determine if the *next* stage should be unlocked
		unlockNext = allQuestionsPassed
	}

	c.JSON(http.StatusOK, response)
}

// ------------------- Admin CRUD -------------------

func GetStages(c *gin.Context) {
	var stages []models.Stage
	database.DB.Preload("Questions").Find(&stages)
	c.JSON(http.StatusOK, stages)
}

func CreateStage(c *gin.Context) {
	var stage models.Stage
	if err := c.ShouldBindJSON(&stage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Create(&stage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stage"})
		return
	}
	c.JSON(http.StatusCreated, stage)
}

func UpdateStage(c *gin.Context) {
	id := c.Param("id")
	var stage models.Stage
	if err := database.DB.First(&stage, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stage not found"})
		return
	}

	if err := c.ShouldBindJSON(&stage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Save(&stage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stage"})
		return
	}

	c.JSON(http.StatusOK, stage)
}

func DeleteStage(c *gin.Context) {
	id := c.Param("id")
	var stage models.Stage
	if err := database.DB.First(&stage, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stage not found"})
		return
	}

	if err := database.DB.Delete(&stage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete stage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stage deleted"})
}

type StageQuestionInput struct {
	Title         string `json:"title" binding:"required"`
	Description   string `json:"description" binding:"required"`
	Type          string `json:"type" binding:"required"`
	RightSolution string `json:"right_solution" binding:"required"`
	Test          string `json:"test" binding:"required"`
	OrderIndex    int    `json:"order_index"`
}

func CreateStageQuestion(c *gin.Context) {
	stageID := c.Param("id")
	var input StageQuestionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var stage models.Stage
	if err := database.DB.First(&stage, stageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stage not found"})
		return
	}

	q := models.StageQuestion{
		StageID:       stage.ID,
		Title:         input.Title,
		Description:   input.Description,
		Type:          input.Type,
		RightSolution: input.RightSolution,
		Test:          input.Test,
		OrderIndex:    input.OrderIndex,
	}

	if err := database.DB.Create(&q).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question"})
		return
	}
	c.JSON(http.StatusCreated, q)
}

func UpdateStageQuestion(c *gin.Context) {
	qID := c.Param("qid")
	var q models.StageQuestion
	if err := database.DB.First(&q, qID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	var input StageQuestionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	q.Title = input.Title
	q.Description = input.Description
	q.Type = input.Type
	q.RightSolution = input.RightSolution
	q.Test = input.Test
	q.OrderIndex = input.OrderIndex

	database.DB.Save(&q)
	c.JSON(http.StatusOK, q)
}

func DeleteStageQuestion(c *gin.Context) {
	qID := c.Param("qid")
	database.DB.Where("stage_question_id = ?", qID).Delete(&models.Submission{}) // Cleanup
	database.DB.Delete(&models.StageQuestion{}, qID)
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func GetStageQuestionByID(c *gin.Context) {
	qID := c.Param("id")
	var q models.StageQuestion
	if err := database.DB.First(&q, qID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":          q.ID,
		"title":       q.Title,
		"description": q.Description,
		"type":        q.Type,
		"stage_id":    q.StageID,
		"order_index": q.OrderIndex,
	})
}

package database

import (
	"backend/models"
	"log"

	"gorm.io/gorm"
)

func seedStagesAndExercises(db *gorm.DB) {
	stagesDef := getStages()
	stages := make([]models.Stage, len(stagesDef))

	for i, s := range stagesDef {
		var existing models.Stage
		if err := db.Where("order_index = ?", s.OrderIndex).First(&existing).Error; err != nil {
			db.Create(&s)
			stages[i] = s
		} else {
			existing.Title = s.Title
			existing.Description = s.Description
			db.Save(&existing)
			stages[i] = existing
		}
	}

	// Load all questions from the questions/ directory
	questions := loadAllQuestions(stages)

	for _, q := range questions {
		var existing models.StageQuestion
		if err := db.Where("stage_id = ? AND order_index = ?", q.StageID, q.OrderIndex).First(&existing).Error; err != nil {
			db.Create(&q)
		} else {
			existing.Title = q.Title
			existing.Description = q.Description
			existing.Type = q.Type
			existing.RightSolution = q.RightSolution
			existing.Test = q.Test
			db.Save(&existing)
		}
	}
	log.Println("Successfully seeded English learning path stages and questions.")
}

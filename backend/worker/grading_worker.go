package worker

import (
	"log"
	"time"

	"backend/config"
	"backend/database"
	"backend/models"
	"backend/sandbox"
)

var (
	// SubmissionQueue holds IDs of submissions waiting to be graded.
	// A buffered channel of size 100 handles a backlog natively.
	SubmissionQueue = make(chan uint, 100)
)

// StartGradingWorker begins background worker goroutines that process submissions concurrently
func StartGradingWorker() {
	workers := config.AppConfig.SandboxWorkers
	if workers <= 0 {
		workers = 3
	}

	log.Printf("Grading Worker Pool started with %d workers listening for submissions...", workers)
	for i := 1; i <= workers; i++ {
		workerID := i
		go func(wID int) {
			log.Printf("Grading Worker #%d ready.", wID)
			for submissionID := range SubmissionQueue {
				processSubmission(submissionID)
			}
		}(workerID)
	}
}

func processSubmission(id uint) {
	// 1. Fetch the submission from the Database
	var submission models.Submission
	if err := database.DB.First(&submission, id).Error; err != nil {
		log.Printf("Worker: Submission %d not found", id)
		return
	}

	// 2. Mark as Running
	database.DB.Model(&submission).Update("status", "Running")

	// 3. Fetch exercise or stage question details
	var rightSolution string
	var testCase string

	if submission.ExerciseID != nil {
		var exercise models.Exercise
		if err := database.DB.First(&exercise, *submission.ExerciseID).Error; err != nil {
			database.DB.Model(&submission).Updates(models.Submission{
				Status:       "Error",
				ErrorMessage: "Failed to load exercise details.",
			})
			return
		}
		rightSolution = exercise.RightSolution
		testCase = exercise.Test
	} else if submission.StageQuestionID != nil {
		var stageQ models.StageQuestion
		if err := database.DB.First(&stageQ, *submission.StageQuestionID).Error; err != nil {
			database.DB.Model(&submission).Updates(models.Submission{
				Status:       "Error",
				ErrorMessage: "Failed to load stage question details.",
			})
			return
		}
		rightSolution = stageQ.RightSolution
		testCase = stageQ.Test
	} else {
		database.DB.Model(&submission).Updates(models.Submission{
			Status:       "Error",
			ErrorMessage: "Invalid submission request.",
		})
		return
	}

	// 4. Run the Sandbox
	log.Printf("Worker: Running Sandbox for Submission %d", id)
	start := time.Now()

	result := sandbox.RunSubmission(submission.Code, rightSolution, testCase)

	log.Printf("Worker: Finished Sandbox for Submission %d in %v", id, time.Since(start))

	// 5. Update Database with Results
	if result.Passed {
		database.DB.Model(&submission).Updates(map[string]interface{}{
			"status":          "Passed",
			"error_message":   "",
			"expected_output": "",
			"actual_output":   "",
			"failed_testcase": "",
		})

		// Award XP for first-time pass
		awardXPIfFirstPass(submission)
	} else {
		// Identify if it's a syntax error vs logic failure
		status := "Failed"
		if result.ExpectedOutput == "" && result.ActualOutput == "" && result.ErrorMessage != "" {
			// Typical syntax or runtime total crash without outputs to compare
			status = "SyntaxError"
		}

		database.DB.Model(&submission).Updates(map[string]interface{}{
			"status":          status,
			"error_message":   result.ErrorMessage,
			"expected_output": result.ExpectedOutput,
			"actual_output":   result.ActualOutput,
			// result.ErrorMessage contains the failed testcase argument if it was a logic error
			"failed_testcase": extractFailedTestcaseFromMsg(result.ErrorMessage),
		})
	}
}

// awardXPIfFirstPass checks if the user has already passed this specific question/exercise before.
// If this is their first pass, it awards XP and recalculates their level.
func awardXPIfFirstPass(submission models.Submission) {
	var previousPasses int64

	if submission.StageQuestionID != nil {
		database.DB.Model(&models.Submission{}).
			Where("user_id = ? AND stage_question_id = ? AND status = ? AND id != ?",
				submission.UserID, *submission.StageQuestionID, "Passed", submission.ID).
			Count(&previousPasses)
	} else if submission.ExerciseID != nil {
		database.DB.Model(&models.Submission{}).
			Where("user_id = ? AND exercise_id = ? AND status = ? AND id != ?",
				submission.UserID, *submission.ExerciseID, "Passed", submission.ID).
			Count(&previousPasses)
	}

	if previousPasses > 0 {
		log.Printf("Worker: User %d already passed this question/exercise. No XP awarded.", submission.UserID)
		return
	}

	// Determine XP reward
	xpReward := 30 // Default for regular exercises
	if submission.StageQuestionID != nil {
		xpReward = 50 // Stage questions are worth more
	}

	// Fetch user, add XP, recalculate level
	var user models.User
	if err := database.DB.First(&user, submission.UserID).Error; err != nil {
		log.Printf("Worker: Failed to fetch user %d for XP award: %v", submission.UserID, err)
		return
	}

	newXP := user.XP + xpReward
	newLevel := calculateLevel(newXP)

	database.DB.Model(&user).Updates(map[string]interface{}{
		"xp":    newXP,
		"level": newLevel,
	})

	log.Printf("Worker: Awarded %d XP to user %d. Total XP: %d, Level: %d", xpReward, submission.UserID, newXP, newLevel)
}

// calculateLevel returns the level for a given XP amount
func calculateLevel(xp int) int {
	thresholds := []int{0, 100, 250, 500, 800, 1200, 1800, 2500, 3500, 5000}
	level := 1
	for i, threshold := range thresholds {
		if xp >= threshold {
			level = i + 1
		}
	}
	return level
}

// Helper to clean up error messages if needed
func extractFailedTestcaseFromMsg(msg string) string {
	// Simple passthrough for now, since sandbox sets the message
	return msg
}

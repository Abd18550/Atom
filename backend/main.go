package main

import (
	"log"
	"time"

	"backend/config"
	"backend/controllers"
	"backend/database"
	"backend/middleware"
	"backend/worker"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from .env / environment
	config.Load()

	// Set Gin mode from config
	gin.SetMode(config.AppConfig.GinMode)

	// Initialize database
	database.ConnectDB()

	r := gin.Default()

	// CORS Setup — configurable origins from environment
	r.Use(cors.New(cors.Config{
		AllowOrigins:     config.AppConfig.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Public routes
	api := r.Group("/api")
	{
		api.POST("/login", controllers.Login)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/me", controllers.GetMe)
		protected.PUT("/profile", controllers.UpdateProfile)
		protected.PUT("/theme", controllers.UpdateTheme)
		protected.GET("/student-stats", controllers.GetStudentStats)
		protected.GET("/student/group-comparison", controllers.GetStudentGroupComparison)
		protected.GET("/mentor/analytics", middleware.RoleMiddleware("Admin", "Supervisor", "Mentor"), controllers.GetMentorAnalytics)

		// Admin, Supervisor, & Mentor can access the basic user list
		usersGroup := protected.Group("/users")
		usersGroup.Use(middleware.RoleMiddleware("Admin", "Supervisor", "Mentor"))
		{
			usersGroup.GET("", controllers.GetUsers)
			usersGroup.POST("", controllers.CreateUser)
			usersGroup.DELETE("/:id", controllers.DeleteUser)
		}

		groupsGroup := protected.Group("/groups")
		groupsGroup.Use(middleware.RoleMiddleware("Admin", "Supervisor", "Mentor"))
		{
			groupsGroup.GET("", controllers.GetGroups)
			groupsGroup.GET("/:id", controllers.GetGroupByID)
			groupsGroup.POST("", controllers.CreateGroup)
			groupsGroup.PUT("/:id", controllers.UpdateGroup)
			groupsGroup.DELETE("/:id", controllers.DeleteGroup)
			groupsGroup.POST("/assign-student", controllers.AssignStudentToGroup)
		}

		// Learning Path & Stages
		protected.GET("/learning-path", controllers.GetLearningPath)

		stagesGroup := protected.Group("/stages")
		{
			stagesGroup.GET("", controllers.GetStages)
			stagesGroup.POST("", middleware.RoleMiddleware("Admin", "Supervisor"), controllers.CreateStage)
			stagesGroup.PUT("/:id", middleware.RoleMiddleware("Admin", "Supervisor"), controllers.UpdateStage)
			stagesGroup.DELETE("/:id", middleware.RoleMiddleware("Admin", "Supervisor"), controllers.DeleteStage)

			stagesGroup.POST("/:id/questions", middleware.RoleMiddleware("Admin", "Supervisor"), controllers.CreateStageQuestion)
			stagesGroup.PUT("/:id/questions/:qid", middleware.RoleMiddleware("Admin", "Supervisor"), controllers.UpdateStageQuestion)
			stagesGroup.DELETE("/:id/questions/:qid", middleware.RoleMiddleware("Admin", "Supervisor"), controllers.DeleteStageQuestion)
		}

		// Publicly accessible route to get a single stage question (no secrets returned)
		protected.GET("/stage-questions/:id", controllers.GetStageQuestionByID)

		// Grading & Exercise System
		exercisesGroup := protected.Group("/exercises")
		{
			exercisesGroup.GET("", controllers.GetExercises)
			exercisesGroup.GET("/:id", controllers.GetExerciseByID)
			exercisesGroup.POST("", middleware.RoleMiddleware("Admin", "Supervisor"), controllers.CreateExercise)
			exercisesGroup.PUT("/:id", middleware.RoleMiddleware("Admin", "Supervisor"), controllers.UpdateExercise)
			exercisesGroup.DELETE("/:id", middleware.RoleMiddleware("Admin", "Supervisor"), controllers.DeleteExercise)
		}

		submissionsGroup := protected.Group("/submissions")
		{
			// Rate limit: 1 submission per 5 seconds per user
			submissionsGroup.POST("", middleware.RateLimitMiddleware(5*time.Second), controllers.SubmitCode)
			submissionsGroup.GET("/:id", controllers.GetSubmissionStatus)
		}
	}

	// Start Background Workers
	worker.StartGradingWorker()

	log.Printf("Server running on port %s...", config.AppConfig.Port)
	r.Run(":" + config.AppConfig.Port)
}

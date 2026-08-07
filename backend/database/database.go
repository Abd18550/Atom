package database

import (
	"log"
	"time"

	"backend/config"
	"backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	database, err := gorm.Open(postgres.Open(config.AppConfig.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool for concurrent access
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatal("Failed to get underlying DB connection:", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	err = database.AutoMigrate(
		&models.User{},
		&models.StudentGroup{},
		&models.Stage{},
		&models.StageQuestion{},
		&models.Exercise{},
		&models.Submission{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	DB = database

	// Run seeders
	seedStagesAndExercises(DB)
	seedAdmin()
}

func seedAdmin() {
	var admin models.User
	result := DB.Where("username = ?", config.AppConfig.AdminUsername).First(&admin)

	if result.Error == gorm.ErrRecordNotFound {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(config.AppConfig.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Failed to hash admin password")
		}

		adminUser := models.User{
			Username:     config.AppConfig.AdminUsername,
			Email:        config.AppConfig.AdminEmail,
			PasswordHash: string(passwordHash),
			Role:         "Admin",
			FullName:     config.AppConfig.AdminFullName,
		}

		DB.Create(&adminUser)
		log.Println("Admin user seeded successfully.")
	} else {
		log.Println("Admin user already exists.")
	}
}

// ResetAndSeedDB wipes submissions, student groups, and student users to reset the DB to clean state
func ResetAndSeedDB() error {
	// Truncate submissions and student_groups
	if err := DB.Exec("TRUNCATE TABLE submissions, student_groups CASCADE;").Error; err != nil {
		DB.Exec("DELETE FROM submissions;")
		DB.Exec("DELETE FROM student_groups;")
	}

	// Delete all student users
	if err := DB.Where("role = ?", "Student").Delete(&models.User{}).Error; err != nil {
		return err
	}

	// Re-run seeders
	seedStagesAndExercises(DB)
	seedAdmin()

	return nil
}

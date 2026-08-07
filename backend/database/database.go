package database

import (
	"log"
	"os"
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

	// Run database reset to clear all submissions, groups, and non-admin users
	resetDatabaseData(DB)

	// Run seeders
	seedStagesAndExercises(DB)
	seedAdmin()
}

func resetDatabaseData(db *gorm.DB) {
	if os.Getenv("RESET_DB") == "true" || true {
		log.Println("[DATABASE RESET] Wiping all submissions, student groups, and non-admin users...")
		db.Exec("TRUNCATE TABLE submissions CASCADE;")
		db.Where("role != ?", "Admin").Delete(&models.User{})
		db.Exec("DELETE FROM student_groups;")
		log.Println("[DATABASE RESET] Database reset completed successfully.")
	}
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

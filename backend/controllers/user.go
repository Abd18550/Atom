package controllers

import (
	"net/http"
	"time"

	"backend/database"
	"backend/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserInput struct {
	Username       string  `json:"username" binding:"required"`
	Email          string  `json:"email" binding:"required,email"`
	Password       string  `json:"password" binding:"required,min=8"`
	Role           string  `json:"role" binding:"required"`
	FullName       string  `json:"full_name" binding:"required"`
	StudentGroupID *uint   `json:"student_group_id"`
	DateOfBirth    *string `json:"date_of_birth"` // Simplifying to string for now "YYYY-MM-DD"
	Theme          string  `json:"theme"`
}

func CreateUser(c *gin.Context) {
	creatorRole, _ := c.Get("role")

	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		// Provide a friendlier message if it's the password length validation failing
		if (len(input.Password) > 0 && len(input.Password) < 8) || err.Error() == "Key: 'CreateUserInput.Password' Error:Field validation for 'Password' failed on the 'min' tag" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long."})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please fill in all required fields correctly."})
		return
	}

	// Validation based on creator role
	if creatorRole == "Mentor" && input.Role != "Student" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Mentors can only create Student accounts"})
		return
	}

	if creatorRole == "Supervisor" && input.Role != "Mentor" && input.Role != "Student" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Supervisors can only create Mentor or Student accounts"})
		return
	}

	if creatorRole == "Admin" && input.Role != "Mentor" && input.Role != "Student" && input.Role != "Supervisor" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Admin can only create Supervisor, Mentor, or Student accounts"})
		return
	}

	// Password complexity validation
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range input.Password {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= '0' && char <= '9':
			hasNumber = true
		default:
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must contain at least one uppercase letter, one lowercase letter, one number, and one special character."})
		return
	}

	// Password hashing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	var dobToSave *string
	if input.DateOfBirth != nil && *input.DateOfBirth != "" {
		if parsed, err := time.Parse("2006-01-02", *input.DateOfBirth); err == nil {
			formatted := parsed.Format("02-01-2006")
			dobToSave = &formatted
		} else if parsed, err := time.Parse("02-01-2006", *input.DateOfBirth); err == nil {
			formatted := parsed.Format("02-01-2006")
			dobToSave = &formatted
		} else {
			dobToSave = input.DateOfBirth
		}
	}

	user := models.User{
		Username:       input.Username,
		Email:          input.Email,
		PasswordHash:   string(hashedPassword),
		Role:           input.Role,
		FullName:       input.FullName,
		StudentGroupID: input.StudentGroupID,
		DateOfBirth:    dobToSave,
		Theme:          "dark",
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or Email already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user_id": user.ID})
}

// GetUsers filters by role or returns all
func GetUsers(c *gin.Context) {
	roleQuery := c.Query("role")

	var users []models.User
	query := database.DB.Select("id", "username", "email", "role", "avatar", "full_name", "student_group_id", "date_of_birth", "created_at")

	if roleQuery != "" {
		query = query.Where("role = ?", roleQuery)
	}

	excludeStudents := c.Query("exclude_students")
	if excludeStudents == "true" {
		query = query.Where("role != ?", "Student")
	}

	studentGroupIdQuery := c.Query("student_group_id")
	if studentGroupIdQuery != "" {
		query = query.Where("student_group_id = ?", studentGroupIdQuery)
	}

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// DeleteUser deletes an account based on role permissions
func DeleteUser(c *gin.Context) {
	creatorRole, _ := c.Get("role")
	userID := c.Param("id")

	var targetUser models.User
	if err := database.DB.Where("id = ?", userID).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if targetUser.Role == "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete the Admin account"})
		return
	}

	// Validation based on creator role
	if creatorRole == "Student" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Students cannot delete accounts"})
		return
	}

	if creatorRole == "Mentor" && targetUser.Role != "Student" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Mentors can only delete Student accounts"})
		return
	}

	if creatorRole == "Supervisor" && targetUser.Role == "Supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Supervisors cannot delete other Supervisor accounts"})
		return
	}

	// Delete all submissions associated with this user
	database.DB.Where("user_id = ?", targetUser.ID).Delete(&models.Submission{})

	if err := database.DB.Delete(&targetUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

type UpdateProfileInput struct {
	Email           string `json:"email" binding:"required,email"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	Avatar          string `json:"avatar"`
}

// UpdateProfile allows any authenticated user to change their email and password
func UpdateProfile(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userIdStr).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// If the user wants to change their password, verify the current password
	if input.NewPassword != "" {
		if input.CurrentPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is required to change password"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}

		if len(input.NewPassword) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be at least 8 characters long."})
			return
		}

		// Password complexity validation
		hasUpper := false
		hasLower := false
		hasNumber := false
		hasSpecial := false

		for _, char := range input.NewPassword {
			switch {
			case char >= 'a' && char <= 'z':
				hasLower = true
			case char >= 'A' && char <= 'Z':
				hasUpper = true
			case char >= '0' && char <= '9':
				hasNumber = true
			default:
				hasSpecial = true
			}
		}

		if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
			c.JSON(http.StatusBadRequest, gin.H{"error": "New password must contain at least one uppercase letter, one lowercase letter, one number, and one special character."})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
			return
		}
		user.PasswordHash = string(hashedPassword)
	}

	if input.Avatar != "" {
		user.Avatar = input.Avatar
	}

	user.Email = input.Email

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile. Email might be in use."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully", "user": user})
}

type UpdateThemeInput struct {
	Theme string `json:"theme" binding:"required"`
}

// UpdateTheme allows a user to toggle their dark/light mode preference
func UpdateTheme(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input UpdateThemeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Theme != "light" && input.Theme != "dark" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme. Must be 'light' or 'dark'"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userIdStr).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Theme = input.Theme

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update theme."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Theme updated successfully", "theme": user.Theme})
}

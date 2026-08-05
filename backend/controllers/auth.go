package controllers

import (
	"net/http"

	"backend/database"
	"backend/models"
	"backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ? OR email = ?", input.Login, input.Login).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login credentials"})
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":               user.ID,
			"username":         user.Username,
			"email":            user.Email,
			"role":             user.Role,
			"avatar":           user.Avatar,
			"theme":            user.Theme,
			"full_name":        user.FullName,
			"student_group_id": user.StudentGroupID,
			"date_of_birth":    user.DateOfBirth,
			"xp":               user.XP,
			"level":            user.Level,
		},
	})
}

func GetMe(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               user.ID,
		"username":         user.Username,
		"email":            user.Email,
		"role":             user.Role,
		"avatar":           user.Avatar,
		"theme":            user.Theme,
		"full_name":        user.FullName,
		"student_group_id": user.StudentGroupID,
		"date_of_birth":    user.DateOfBirth,
		"xp":               user.XP,
		"level":            user.Level,
	})
}

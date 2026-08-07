package controllers

import (
	"net/http"

	"backend/database"
	"backend/models"

	"github.com/gin-gonic/gin"
)

type CreateGroupInput struct {
	SchoolName   string `json:"school_name" binding:"required"`
	Class        string `json:"class" binding:"required"`
	AcademicYear string `json:"academic_year" binding:"required"`
}

func CreateGroup(c *gin.Context) {
	creatorRole, _ := c.Get("role")
	creatorID, _ := c.Get("userID")

	if creatorRole == "Student" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Students cannot create groups"})
		return
	}

	var input CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := models.StudentGroup{
		SchoolName:   input.SchoolName,
		Class:        input.Class,
		AcademicYear: input.AcademicYear,
		CreatedByID:  creatorID.(uint),
	}

	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Group created successfully",
		"group_id": group.ID,
	})
}

// GetGroups returns groups filtered by ownership:
// - Mentors only see groups they created
// - Admins/Supervisors see all groups
func GetGroups(c *gin.Context) {
	role, _ := c.Get("role")
	userID, _ := c.Get("userID")

	var groups []models.StudentGroup

	query := database.DB
	if role == "Mentor" {
		query = query.Where("created_by_id = ?", userID)
	}

	if err := query.Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch groups"})
		return
	}

	c.JSON(http.StatusOK, groups)
}

func GetGroupByID(c *gin.Context) {
	groupID := c.Param("id")
	role, _ := c.Get("role")
	userID, _ := c.Get("userID")

	var group models.StudentGroup
	if err := database.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Mentors can only view groups they created
	if role == "Mentor" && group.CreatedByID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this group"})
		return
	}

	c.JSON(http.StatusOK, group)
}

func DeleteGroup(c *gin.Context) {
	groupID := c.Param("id")
	role, _ := c.Get("role")
	userID, _ := c.Get("userID")

	var group models.StudentGroup
	if err := database.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Mentors can only delete groups they created
	if role == "Mentor" && group.CreatedByID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete groups you created"})
		return
	}

	// Unassign students from the group first
	database.DB.Model(&models.User{}).Where("student_group_id = ?", group.ID).Update("student_group_id", nil)

	if err := database.DB.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

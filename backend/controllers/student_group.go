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

func GetGroups(c *gin.Context) {
	var groups []models.StudentGroup

	if err := database.DB.Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch groups"})
		return
	}

	c.JSON(http.StatusOK, groups)
}

func GetGroupByID(c *gin.Context) {
	groupID := c.Param("id")

	var group models.StudentGroup
	if err := database.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, group)
}

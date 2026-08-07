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

	var creatorID *uint
	if uid, exists := c.Get("userID"); exists {
		if id, ok := uid.(uint); ok {
			creatorID = &id
		} else if idFloat, ok := uid.(float64); ok {
			idUint := uint(idFloat)
			creatorID = &idUint
		}
	}

	group := models.StudentGroup{
		SchoolName:      input.SchoolName,
		Class:           input.Class,
		AcademicYear:    input.AcademicYear,
		CreatedByUserID: creatorID,
	}

	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	database.DB.Preload("CreatedBy").First(&group, group.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Group created successfully",
		"group_id": group.ID,
		"group":    group,
	})
}

func GetGroups(c *gin.Context) {
	creatorRole, _ := c.Get("role")
	var groups []models.StudentGroup

	if creatorRole == "Admin" || creatorRole == "Supervisor" {
		if err := database.DB.Preload("CreatedBy").Find(&groups).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch groups"})
			return
		}
	} else { // Mentor
		var mentorID uint
		if uid, exists := c.Get("userID"); exists {
			if id, ok := uid.(uint); ok {
				mentorID = id
			} else if idFloat, ok := uid.(float64); ok {
				mentorID = uint(idFloat)
			}
		}

		if err := database.DB.Preload("CreatedBy").Where("created_by_user_id = ?", mentorID).Find(&groups).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch mentor groups"})
			return
		}
	}

	c.JSON(http.StatusOK, groups)
}

func GetGroupByID(c *gin.Context) {
	groupID := c.Param("id")

	var group models.StudentGroup
	if err := database.DB.Preload("CreatedBy").First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, group)
}

func UpdateGroup(c *gin.Context) {
	groupID := c.Param("id")
	creatorRole, _ := c.Get("role")

	var group models.StudentGroup
	if err := database.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Check ownership if mentor
	if creatorRole != "Admin" && creatorRole != "Supervisor" {
		var mentorID uint
		if uid, exists := c.Get("userID"); exists {
			if id, ok := uid.(uint); ok {
				mentorID = id
			} else if idFloat, ok := uid.(float64); ok {
				mentorID = uint(idFloat)
			}
		}

		if group.CreatedByUserID == nil || *group.CreatedByUserID != mentorID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to edit this group"})
			return
		}
	}

	var input CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group.SchoolName = input.SchoolName
	group.Class = input.Class
	group.AcademicYear = input.AcademicYear

	if err := database.DB.Save(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	database.DB.Preload("CreatedBy").First(&group, group.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Group updated successfully", "group": group})
}

func DeleteGroup(c *gin.Context) {
	groupID := c.Param("id")
	creatorRole, _ := c.Get("role")

	var group models.StudentGroup
	if err := database.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Check ownership if mentor
	if creatorRole != "Admin" && creatorRole != "Supervisor" {
		var mentorID uint
		if uid, exists := c.Get("userID"); exists {
			if id, ok := uid.(uint); ok {
				mentorID = id
			} else if idFloat, ok := uid.(float64); ok {
				mentorID = uint(idFloat)
			}
		}

		if group.CreatedByUserID == nil || *group.CreatedByUserID != mentorID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this group"})
			return
		}
	}

	// Unassign students linked to this group
	database.DB.Model(&models.User{}).Where("student_group_id = ?", group.ID).Update("student_group_id", nil)

	if err := database.DB.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

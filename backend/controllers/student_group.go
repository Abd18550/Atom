package controllers

import (
	"log"
	"net/http"
	"strconv"

	"backend/database"
	"backend/models"

	"github.com/gin-gonic/gin"
)

type CreateGroupInput struct {
	SchoolName   string `json:"school_name" binding:"required"`
	Class        string `json:"class" binding:"required"`
	AcademicYear string `json:"academic_year" binding:"required"`
}

type AssignStudentInput struct {
	UserID         uint  `json:"user_id" binding:"required"`
	StudentGroupID *uint `json:"student_group_id"`
}

func CreateGroup(c *gin.Context) {
	creatorRole, _ := c.Get("role")
	userIDVal, exists := c.Get("userID")

	if creatorRole == "Student" || !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only Mentors and Admins can create groups"})
		return
	}

	userID := userIDVal.(uint)

	var input CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := models.StudentGroup{
		SchoolName:   input.SchoolName,
		Class:        input.Class,
		AcademicYear: input.AcademicYear,
		CreatedByID:  userID,
	}

	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Group created successfully", "group": group})
}

// GetGroups filters groups by creator for Mentors or returns all for Admin/Supervisor
func GetGroups(c *gin.Context) {
	creatorRole, _ := c.Get("role")
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	var groups []models.StudentGroup
	query := database.DB.Preload("CreatedBy")

	if creatorRole == "Mentor" {
		query = query.Where("created_by_id = ? OR created_by_id = 0 OR created_by_id IS NULL", userID)
	}

	if err := query.Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	c.JSON(http.StatusOK, groups)
}

func GetGroupByID(c *gin.Context) {
	groupIDStr := c.Param("id")
	idNum, err := strconv.Atoi(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID format"})
		return
	}

	creatorRole, _ := c.Get("role")
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	var group models.StudentGroup
	if err := database.DB.Preload("CreatedBy").Where("id = ?", idNum).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	if creatorRole == "Mentor" && group.CreatedByID != 0 && group.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. You can only view groups you created."})
		return
	}

	c.JSON(http.StatusOK, group)
}

func UpdateGroup(c *gin.Context) {
	groupIDStr := c.Param("id")
	idNum, err := strconv.Atoi(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID format"})
		return
	}

	creatorRole, _ := c.Get("role")
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	var group models.StudentGroup
	if err := database.DB.Where("id = ?", idNum).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	if creatorRole == "Mentor" && group.CreatedByID != 0 && group.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. You can only manage groups you created."})
		return
	}

	var input CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group.SchoolName = input.SchoolName
	group.Class = input.Class
	group.AcademicYear = input.AcademicYear
	if group.CreatedByID == 0 {
		group.CreatedByID = userID
	}

	if err := database.DB.Save(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group updated successfully", "group": group})
}

func DeleteGroup(c *gin.Context) {
	groupIDStr := c.Param("id")
	idNum, err := strconv.Atoi(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID format"})
		return
	}

	creatorRole, _ := c.Get("role")
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	log.Printf("[DeleteGroup] Attempting delete for group ID: %d by user %d (role: %s)", idNum, userID, creatorRole)

	var group models.StudentGroup
	if err := database.DB.Where("id = ?", idNum).First(&group).Error; err != nil {
		log.Printf("[DeleteGroup] Group ID %d not found in DB: %v", idNum, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	if creatorRole == "Mentor" && group.CreatedByID != 0 && group.CreatedByID != userID {
		log.Printf("[DeleteGroup] Forbidden delete for group ID %d created by %d", idNum, group.CreatedByID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. You can only delete groups you created."})
		return
	}

	// 1. Find all student user IDs belonging to this group
	var studentIDs []uint
	database.DB.Model(&models.User{}).Where("student_group_id = ?", group.ID).Pluck("id", &studentIDs)

	// 2. If there are students in this group, delete all their submissions first
	if len(studentIDs) > 0 {
		if err := database.DB.Where("user_id IN (?)", studentIDs).Delete(&models.Submission{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete student submissions for group"})
			return
		}

		// 3. Delete all student users in this group
		if err := database.DB.Where("id IN (?)", studentIDs).Delete(&models.User{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete students in group"})
			return
		}
	}

	// 4. Clear any lingering foreign keys if any non-student was linked
	database.DB.Model(&models.User{}).Where("student_group_id = ?", group.ID).Update("student_group_id", nil)

	// 5. Delete the StudentGroup record itself
	if err := database.DB.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group and all assigned students were deleted successfully"})
}

func AssignStudentToGroup(c *gin.Context) {
	creatorRole, _ := c.Get("role")
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	var input AssignStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var student models.User
	if err := database.DB.First(&student, input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	if student.Role != "Student" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target user is not a student"})
		return
	}

	if input.StudentGroupID != nil && *input.StudentGroupID > 0 {
		var group models.StudentGroup
		if err := database.DB.First(&group, *input.StudentGroupID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Target group not found"})
			return
		}

		if creatorRole == "Mentor" && group.CreatedByID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. You can only assign students to groups you created."})
			return
		}

		student.StudentGroupID = input.StudentGroupID
	} else {
		student.StudentGroupID = nil
	}

	if err := database.DB.Save(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student group assignment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student group updated successfully"})
}

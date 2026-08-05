package controllers

import (
	"math"
	"net/http"
	"time"

	"backend/database"
	"backend/models"

	"github.com/gin-gonic/gin"
)

// GroupAnalyticsItem represents performance statistics for a group
type GroupAnalyticsItem struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	StudentCount int     `json:"student_count"`
	AverageXP    float64 `json:"average_xp"`
	PassedCount  int     `json:"passed_count"`
}

// StudentAnalyticsItem represents detailed performance metrics for a student
type StudentAnalyticsItem struct {
	ID               uint       `json:"id"`
	Username         string     `json:"username"`
	FullName         string     `json:"full_name"`
	Email            string     `json:"email"`
	Avatar           string     `json:"avatar"`
	GroupName        string     `json:"group_name"`
	XP               int        `json:"xp"`
	Level            int        `json:"level"`
	LevelTitle       string     `json:"level_title"`
	PassedQuestions  int        `json:"passed_questions"`
	TotalSubmissions int        `json:"total_submissions"`
	SuccessRate      float64    `json:"success_rate"`
	LastActive       *time.Time `json:"last_active"`
}

// RecentActivityItem represents a recent submission attempt by a student
type RecentActivityItem struct {
	ID            uint      `json:"id"`
	StudentID     uint      `json:"student_id"`
	StudentName   string    `json:"student_name"`
	StudentAvatar string    `json:"student_avatar"`
	TargetTitle   string    `json:"target_title"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// GetMentorAnalytics aggregates overall student performance, group comparison, and recent activity
func GetMentorAnalytics(c *gin.Context) {
	// 1. Fetch all students with group information
	var students []models.User
	if err := database.DB.Preload("StudentGroup").Where("role = ?", "Student").Find(&students).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch student data"})
		return
	}

	// 2. Fetch all student groups
	var groups []models.StudentGroup
	database.DB.Find(&groups)

	// 3. Fetch all submissions to aggregate metrics
	var submissions []models.Submission
	database.DB.Find(&submissions)

	// Fetch stage questions and exercises for activity title lookup
	var stageQuestions []models.StageQuestion
	database.DB.Find(&stageQuestions)
	sqMap := make(map[uint]string)
	for _, sq := range stageQuestions {
		sqMap[sq.ID] = sq.Title
	}

	var exercises []models.Exercise
	database.DB.Find(&exercises)
	exMap := make(map[uint]string)
	for _, ex := range exercises {
		exMap[ex.ID] = ex.Title
	}

	totalStudents := len(students)
	totalSubmissions := len(submissions)
	passedSubmissionsCount := 0

	// Track per-student metrics
	type userStat struct {
		totalSubmissions int
		passedSubmissions int
		passedQuestionsMap map[uint]bool
		lastActive       *time.Time
	}
	userStatsMap := make(map[uint]*userStat)

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	activeStudentsMap := make(map[uint]bool)

	for _, s := range submissions {
		if s.Status == "Passed" {
			passedSubmissionsCount++
		}

		st, exists := userStatsMap[s.UserID]
		if !exists {
			st = &userStat{
				passedQuestionsMap: make(map[uint]bool),
			}
			userStatsMap[s.UserID] = st
		}

		st.totalSubmissions++
		if s.Status == "Passed" {
			st.passedSubmissions++
			if s.StageQuestionID != nil {
				st.passedQuestionsMap[*s.StageQuestionID] = true
			}
		}

		if st.lastActive == nil || s.CreatedAt.After(*st.lastActive) {
			t := s.CreatedAt
			st.lastActive = &t
		}

		if s.CreatedAt.After(sevenDaysAgo) {
			activeStudentsMap[s.UserID] = true
		}
	}

	// Calculate Overall Pass Rate
	overallPassRate := 0.0
	if totalSubmissions > 0 {
		overallPassRate = math.Round((float64(passedSubmissionsCount)/float64(totalSubmissions))*1000) / 10
	}

	// 4. Build Student Roster Analytics
	studentRoster := make([]StudentAnalyticsItem, 0, len(students))
	groupXPMap := make(map[uint]int)
	groupStudentCountMap := make(map[uint]int)
	groupPassedCountMap := make(map[uint]int)

	for _, s := range students {
		groupName := "Unassigned"
		if s.StudentGroupID != nil && s.StudentGroup != nil {
			groupName = s.StudentGroup.SchoolName + " - " + s.StudentGroup.Class
			groupXPMap[*s.StudentGroupID] += s.XP
			groupStudentCountMap[*s.StudentGroupID]++
		}

		st := userStatsMap[s.ID]
		passedQCount := 0
		totalSub := 0
		passSub := 0
		var lastAct *time.Time

		if st != nil {
			passedQCount = len(st.passedQuestionsMap)
			totalSub = st.totalSubmissions
			passSub = st.passedSubmissions
			lastAct = st.lastActive
		}

		successRate := 0.0
		if totalSub > 0 {
			successRate = math.Round((float64(passSub)/float64(totalSub))*1000) / 10
		}

		if s.StudentGroupID != nil {
			groupPassedCountMap[*s.StudentGroupID] += passedQCount
		}

		studentRoster = append(studentRoster, StudentAnalyticsItem{
			ID:               s.ID,
			Username:         s.Username,
			FullName:         s.FullName,
			Email:            s.Email,
			Avatar:           s.Avatar,
			GroupName:        groupName,
			XP:               s.XP,
			Level:            s.Level,
			LevelTitle:       getLevelTitle(s.Level),
			PassedQuestions:  passedQCount,
			TotalSubmissions: totalSub,
			SuccessRate:      successRate,
			LastActive:       lastAct,
		})
	}

	// 5. Build Group Analytics
	groupAnalyticsList := make([]GroupAnalyticsItem, 0, len(groups))
	for _, g := range groups {
		cnt := groupStudentCountMap[g.ID]
		avgXP := 0.0
		if cnt > 0 {
			avgXP = math.Round((float64(groupXPMap[g.ID])/float64(cnt))*10) / 10
		}
		groupAnalyticsList = append(groupAnalyticsList, GroupAnalyticsItem{
			ID:           g.ID,
			Name:         g.SchoolName + " - " + g.Class,
			StudentCount: cnt,
			AverageXP:    avgXP,
			PassedCount:  groupPassedCountMap[g.ID],
		})
	}

	// 6. Build Recent Activity Stream (Latest 15 submissions)
	var recentSubs []models.Submission
	database.DB.Order("created_at desc").Limit(15).Find(&recentSubs)

	userMap := make(map[uint]models.User)
	for _, s := range students {
		userMap[s.ID] = s
	}

	activityFeed := make([]RecentActivityItem, 0, len(recentSubs))
	for _, sub := range recentSubs {
		u := userMap[sub.UserID]
		title := "Coding Exercise"
		if sub.StageQuestionID != nil && sqMap[*sub.StageQuestionID] != "" {
			title = sqMap[*sub.StageQuestionID]
		} else if sub.ExerciseID != nil && exMap[*sub.ExerciseID] != "" {
			title = exMap[*sub.ExerciseID]
		}

		activityFeed = append(activityFeed, RecentActivityItem{
			ID:            sub.ID,
			StudentID:     sub.UserID,
			StudentName:   u.FullName,
			StudentAvatar: u.Avatar,
			TargetTitle:   title,
			Status:        sub.Status,
			CreatedAt:     sub.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total_students":    totalStudents,
		"active_students":   len(activeStudentsMap),
		"total_submissions": totalSubmissions,
		"overall_pass_rate": overallPassRate,
		"groups":            groupAnalyticsList,
		"students":          studentRoster,
		"activity_feed":     activityFeed,
	})
}

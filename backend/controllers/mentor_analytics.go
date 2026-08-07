package controllers

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"backend/database"
	"backend/models"

	"github.com/gin-gonic/gin"
)

// GroupAnalyticsItem represents performance statistics for a group
type GroupAnalyticsItem struct {
	ID                   uint    `json:"id"`
	SchoolName           string  `json:"school_name"`
	Class                string  `json:"class"`
	AcademicYear         string  `json:"academic_year"`
	Name                 string  `json:"name"`
	CreatedByName        string  `json:"created_by_name"`
	CreatedByUserID      *uint   `json:"created_by_user_id"`
	StudentCount         int     `json:"student_count"`
	AverageXP            float64 `json:"average_xp"`
	TotalPassedQuestions int     `json:"total_passed_questions"`
	AveragePassRate      float64 `json:"average_pass_rate"`
}

// StudentAnalyticsItem represents detailed performance metrics for a student
type StudentAnalyticsItem struct {
	ID               uint       `json:"id"`
	Username         string     `json:"username"`
	FullName         string     `json:"full_name"`
	Email            string     `json:"email"`
	Avatar           string     `json:"avatar"`
	GroupID          *uint      `json:"group_id"`
	GroupName        string     `json:"group_name"`
	XP               int        `json:"xp"`
	Level            int        `json:"level"`
	LevelTitle       string     `json:"level_title"`
	PassedQuestions  int        `json:"passed_questions"`
	TotalSubmissions int        `json:"total_submissions"`
	SuccessRate      float64    `json:"success_rate"`
	LastActive       *time.Time `json:"last_active"`
	GroupAvgXP       float64    `json:"group_avg_xp"`
	GroupAvgPassed   float64    `json:"group_avg_passed"`
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
	requesterRole, _ := c.Get("role")

	var requesterID uint
	if uid, exists := c.Get("userID"); exists {
		if id, ok := uid.(uint); ok {
			requesterID = id
		} else if idFloat, ok := uid.(float64); ok {
			requesterID = uint(idFloat)
		}
	}

	// 1. Fetch groups based on role
	var groups []models.StudentGroup
	if requesterRole == "Admin" || requesterRole == "Supervisor" {
		database.DB.Preload("CreatedBy").Find(&groups)
	} else { // Mentor: strictly groups created by this mentor
		database.DB.Preload("CreatedBy").Where("created_by_user_id = ?", requesterID).Find(&groups)
	}

	groupIDs := make([]uint, 0, len(groups))
	groupIDSet := make(map[uint]bool)
	for _, g := range groups {
		groupIDs = append(groupIDs, g.ID)
		groupIDSet[g.ID] = true
	}

	// 2. Fetch students (STRICTLY role = 'Student')
	var students []models.User
	query := database.DB.Preload("StudentGroup").Where("role = ?", "Student")

	if requesterRole == "Mentor" {
		if len(groupIDs) > 0 {
			query = query.Where("student_group_id IN ?", groupIDs)
		} else {
			// Mentor has no groups yet, so no students returned
			query = query.Where("1 = 0")
		}
	}
	if err := query.Find(&students).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch student data"})
		return
	}

	studentIDSet := make(map[uint]bool)
	studentUserMap := make(map[uint]models.User)
	for _, s := range students {
		studentIDSet[s.ID] = true
		studentUserMap[s.ID] = s
	}

	// 3. Fetch submissions strictly for these students
	var submissions []models.Submission
	if len(studentIDSet) > 0 {
		studentIDs := make([]uint, 0, len(studentIDSet))
		for id := range studentIDSet {
			studentIDs = append(studentIDs, id)
		}
		database.DB.Where("user_id IN ?", studentIDs).Find(&submissions)
	}

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
		totalSubmissions   int
		passedSubmissions  int
		passedQuestionsMap map[uint]bool
		lastActive         *time.Time
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

	// Overall Pass Rate for students
	overallPassRate := 0.0
	if totalSubmissions > 0 {
		overallPassRate = math.Round((float64(passedSubmissionsCount)/float64(totalSubmissions))*1000) / 10
	}

	// Aggregates per group
	groupXPMap := make(map[uint]int)
	groupStudentCountMap := make(map[uint]int)
	groupPassedCountMap := make(map[uint]int)
	groupTotalSubmissionsMap := make(map[uint]int)
	groupPassedSubmissionsMap := make(map[uint]int)

	for _, s := range students {
		if s.StudentGroupID != nil && groupIDSet[*s.StudentGroupID] {
			gid := *s.StudentGroupID
			groupXPMap[gid] += s.XP
			groupStudentCountMap[gid]++

			st := userStatsMap[s.ID]
			if st != nil {
				groupPassedCountMap[gid] += len(st.passedQuestionsMap)
				groupTotalSubmissionsMap[gid] += st.totalSubmissions
				groupPassedSubmissionsMap[gid] += st.passedSubmissions
			}
		}
	}

	// 4. Build Group Analytics List
	groupAnalyticsList := make([]GroupAnalyticsItem, 0, len(groups))
	groupAvgXPMap := make(map[uint]float64)
	groupAvgPassedMap := make(map[uint]float64)

	for _, g := range groups {
		cnt := groupStudentCountMap[g.ID]
		avgXP := 0.0
		avgPassed := 0.0
		passRate := 0.0

		if cnt > 0 {
			avgXP = math.Round((float64(groupXPMap[g.ID])/float64(cnt))*10) / 10
			avgPassed = math.Round((float64(groupPassedCountMap[g.ID])/float64(cnt))*10) / 10
		}
		if groupTotalSubmissionsMap[g.ID] > 0 {
			passRate = math.Round((float64(groupPassedSubmissionsMap[g.ID])/float64(groupTotalSubmissionsMap[g.ID]))*1000) / 10
		}

		groupAvgXPMap[g.ID] = avgXP
		groupAvgPassedMap[g.ID] = avgPassed

		createdByName := "System Admin"
		if g.CreatedBy != nil {
			if g.CreatedBy.FullName != "" {
				createdByName = g.CreatedBy.FullName
			} else {
				createdByName = g.CreatedBy.Username
			}
		}

		groupAnalyticsList = append(groupAnalyticsList, GroupAnalyticsItem{
			ID:                   g.ID,
			SchoolName:           g.SchoolName,
			Class:                g.Class,
			AcademicYear:         g.AcademicYear,
			Name:                 fmt.Sprintf("%s - %s", g.SchoolName, g.Class),
			CreatedByName:        createdByName,
			CreatedByUserID:      g.CreatedByUserID,
			StudentCount:         cnt,
			AverageXP:            avgXP,
			TotalPassedQuestions: groupPassedCountMap[g.ID],
			AveragePassRate:      passRate,
		})
	}

	// 5. Build Student Roster Analytics
	studentRoster := make([]StudentAnalyticsItem, 0, len(students))
	for _, s := range students {
		groupName := "Unassigned"
		var gID *uint
		var gAvgXP float64 = 0.0
		var gAvgPassed float64 = 0.0

		if s.StudentGroupID != nil && s.StudentGroup != nil {
			gID = s.StudentGroupID
			groupName = fmt.Sprintf("%s - %s", s.StudentGroup.SchoolName, s.StudentGroup.Class)
			gAvgXP = groupAvgXPMap[*s.StudentGroupID]
			gAvgPassed = groupAvgPassedMap[*s.StudentGroupID]
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

		studentRoster = append(studentRoster, StudentAnalyticsItem{
			ID:               s.ID,
			Username:         s.Username,
			FullName:         s.FullName,
			Email:            s.Email,
			Avatar:           s.Avatar,
			GroupID:          gID,
			GroupName:        groupName,
			XP:               s.XP,
			Level:            s.Level,
			LevelTitle:       getLevelTitle(s.Level),
			PassedQuestions:  passedQCount,
			TotalSubmissions: totalSub,
			SuccessRate:      successRate,
			LastActive:       lastAct,
			GroupAvgXP:       gAvgXP,
			GroupAvgPassed:   gAvgPassed,
		})
	}

	// 6. Build Recent Activity Feed
	var recentSubs []models.Submission
	if len(studentIDSet) > 0 {
		studentIDs := make([]uint, 0, len(studentIDSet))
		for id := range studentIDSet {
			studentIDs = append(studentIDs, id)
		}
		database.DB.Where("user_id IN ?", studentIDs).Order("created_at desc").Limit(15).Find(&recentSubs)
	}

	activityFeed := make([]RecentActivityItem, 0, len(recentSubs))
	for _, sub := range recentSubs {
		u := studentUserMap[sub.UserID]
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

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
	GroupID          *uint      `json:"group_id"`
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

// getOwnedGroupIDs returns group IDs the current user can access
func getOwnedGroupIDs(c *gin.Context) []uint {
	role, _ := c.Get("role")
	userID, _ := c.Get("userID")

	var groups []models.StudentGroup

	if role == "Mentor" {
		database.DB.Where("created_by_id = ?", userID).Find(&groups)
	} else {
		// Admin / Supervisor see all
		database.DB.Find(&groups)
	}

	ids := make([]uint, len(groups))
	for i, g := range groups {
		ids[i] = g.ID
	}
	return ids
}

// GetMentorAnalytics aggregates student performance scoped to owned groups
func GetMentorAnalytics(c *gin.Context) {
	role, _ := c.Get("role")
	userID, _ := c.Get("userID")

	// 1. Fetch groups scoped to ownership
	var groups []models.StudentGroup
	if role == "Mentor" {
		database.DB.Where("created_by_id = ?", userID).Find(&groups)
	} else {
		database.DB.Find(&groups)
	}

	groupIDSet := make(map[uint]bool)
	for _, g := range groups {
		groupIDSet[g.ID] = true
	}

	// 2. Fetch ONLY students (role=Student) in owned groups
	var allStudents []models.User
	database.DB.Preload("StudentGroup").Where("role = ?", "Student").Find(&allStudents)

	// Filter to only students belonging to owned groups
	var students []models.User
	for _, s := range allStudents {
		if s.StudentGroupID != nil && groupIDSet[*s.StudentGroupID] {
			students = append(students, s)
		}
	}

	// 3. Fetch submissions only for these students
	studentIDs := make([]uint, len(students))
	for i, s := range students {
		studentIDs[i] = s.ID
	}

	var submissions []models.Submission
	if len(studentIDs) > 0 {
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
			GroupID:          s.StudentGroupID,
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

	// 6. Build Recent Activity Stream (Latest 15 submissions from owned-group students)
	var recentSubs []models.Submission
	if len(studentIDs) > 0 {
		database.DB.Where("user_id IN ?", studentIDs).Order("created_at desc").Limit(15).Find(&recentSubs)
	}

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

// GetGroupComparison returns comparison metrics for all owned groups (for bar charts)
func GetGroupComparison(c *gin.Context) {
	role, _ := c.Get("role")
	userID, _ := c.Get("userID")

	var groups []models.StudentGroup
	if role == "Mentor" {
		database.DB.Where("created_by_id = ?", userID).Find(&groups)
	} else {
		database.DB.Find(&groups)
	}

	groupIDSet := make(map[uint]bool)
	for _, g := range groups {
		groupIDSet[g.ID] = true
	}

	// Fetch students (only Students role)
	var allStudents []models.User
	database.DB.Where("role = ?", "Student").Find(&allStudents)

	// Map students by group
	groupStudents := make(map[uint][]models.User)
	for _, s := range allStudents {
		if s.StudentGroupID != nil && groupIDSet[*s.StudentGroupID] {
			groupStudents[*s.StudentGroupID] = append(groupStudents[*s.StudentGroupID], s)
		}
	}

	// Fetch submissions for these students
	var studentIDs []uint
	for _, students := range groupStudents {
		for _, s := range students {
			studentIDs = append(studentIDs, s.ID)
		}
	}

	var submissions []models.Submission
	if len(studentIDs) > 0 {
		database.DB.Where("user_id IN ? AND stage_question_id IS NOT NULL AND status = ?", studentIDs, "Passed").Find(&submissions)
	}

	// Count passed questions per student
	passedPerStudent := make(map[uint]map[uint]bool)
	for _, s := range submissions {
		if _, ok := passedPerStudent[s.UserID]; !ok {
			passedPerStudent[s.UserID] = make(map[uint]bool)
		}
		if s.StageQuestionID != nil {
			passedPerStudent[s.UserID][*s.StageQuestionID] = true
		}
	}

	type GroupComparisonItem struct {
		ID           uint    `json:"id"`
		Name         string  `json:"name"`
		StudentCount int     `json:"student_count"`
		AverageXP    float64 `json:"average_xp"`
		TotalPassed  int     `json:"total_passed"`
		AvgPassed    float64 `json:"avg_passed"`
	}

	result := make([]GroupComparisonItem, 0, len(groups))
	for _, g := range groups {
		students := groupStudents[g.ID]
		count := len(students)
		totalXP := 0
		totalPassed := 0
		for _, s := range students {
			totalXP += s.XP
			if pq, ok := passedPerStudent[s.ID]; ok {
				totalPassed += len(pq)
			}
		}

		avgXP := 0.0
		avgPassed := 0.0
		if count > 0 {
			avgXP = math.Round((float64(totalXP)/float64(count))*10) / 10
			avgPassed = math.Round((float64(totalPassed)/float64(count))*10) / 10
		}

		result = append(result, GroupComparisonItem{
			ID:           g.ID,
			Name:         g.SchoolName + " - " + g.Class,
			StudentCount: count,
			AverageXP:    avgXP,
			TotalPassed:  totalPassed,
			AvgPassed:    avgPassed,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetStudentGroupPeers returns anonymized group peer data for student comparison charts
func GetStudentGroupPeers(c *gin.Context) {
	currentUserID, _ := c.Get("userID")

	// Get current user
	var currentUser models.User
	if err := database.DB.First(&currentUser, currentUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if currentUser.StudentGroupID == nil {
		c.JSON(http.StatusOK, gin.H{"peers": []interface{}{}, "message": "Not assigned to a group"})
		return
	}

	// Get all students in the same group
	var peers []models.User
	database.DB.Where("role = ? AND student_group_id = ?", "Student", *currentUser.StudentGroupID).Find(&peers)

	// Fetch passed questions for each peer
	var peerIDs []uint
	for _, p := range peers {
		peerIDs = append(peerIDs, p.ID)
	}

	var submissions []models.Submission
	if len(peerIDs) > 0 {
		database.DB.Where("user_id IN ? AND stage_question_id IS NOT NULL AND status = ?", peerIDs, "Passed").Find(&submissions)
	}

	passedPerStudent := make(map[uint]map[uint]bool)
	for _, s := range submissions {
		if _, ok := passedPerStudent[s.UserID]; !ok {
			passedPerStudent[s.UserID] = make(map[uint]bool)
		}
		if s.StageQuestionID != nil {
			passedPerStudent[s.UserID][*s.StageQuestionID] = true
		}
	}

	type PeerItem struct {
		Label          string `json:"label"`
		XP             int    `json:"xp"`
		PassedCount    int    `json:"passed_count"`
		IsCurrentUser  bool   `json:"is_current_user"`
	}

	result := make([]PeerItem, 0, len(peers))
	peerIndex := 1
	for _, p := range peers {
		label := ""
		isCurrent := p.ID == currentUserID.(uint)
		if isCurrent {
			label = "You"
		} else {
			label = p.FullName
			if label == "" {
				label = p.Username
			}
			// Abbreviate: "John Doe" -> "John D."
			if len(label) > 12 {
				label = label[:12] + "."
			}
			peerIndex++
		}

		passed := 0
		if pq, ok := passedPerStudent[p.ID]; ok {
			passed = len(pq)
		}

		result = append(result, PeerItem{
			Label:         label,
			XP:            p.XP,
			PassedCount:   passed,
			IsCurrentUser: isCurrent,
		})
	}

	c.JSON(http.StatusOK, gin.H{"peers": result})
}

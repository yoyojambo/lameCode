package app

import (
	"fmt"
	"lameCode/internal/platform/data"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func LoadUserProfileHandlers(r *gin.RouterGroup) {
	r.GET("/user/:username", enableHtmxCache, userProfilePageFunc)
}

type DifficultyStats struct {
	Label    string
	Count    int64
	CssClass string
}

type LanguageStat struct {
	Language        string
	SubmissionCount int64
	AcceptedCount   int64
	AcceptanceRate  float64
}

type CompletedChallenge struct {
	ChallengeID       int64
	Title             string
	Difficulty        int64
	DifficultyLabel   string
	DifficultyClass   string
	Language          string
	CompletedAt       string
	RuntimeInfo       string
	SolutionCreatedAt string
}

type RecentSubmission struct {
	ID               int64
	ChallengeID      int64
	ChallengeTitle   string
	Language         string
	Status           string
	StatusBadgeClass string
	RuntimeInfo      string
	SubmittedAt      string
}

type UserProfileData struct {
	Username               string
	IsAdmin                bool
	MemberSince            string
	DaysActive             int64
	ChallengesCompleted    int64
	TotalSolutions         int64
	AcceptedSolutions      int64
	WrongAnswers           int64
	RuntimeErrors          int64
	SuccessRate            string
	MostUsedLanguage       string
	DifficultyStats        []DifficultyStats
	LanguageStats          []LanguageStat
	CompletedChallenges    []CompletedChallenge
	RecentSubmissions      []RecentSubmission
	HasCompletedChallenges bool
	HasRecentSubmissions   bool
	HasLanguageStats       bool
}

func userProfilePageFunc(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("username is required"))
		return
	}

	repo := data.Repository()

	// Get user stats
	userStats, err := repo.GetUserStats(ctx.Request.Context(), username)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			ctx.HTML(http.StatusNotFound, "error.html", gin.H{
				"message": fmt.Sprintf("User '%s' not found", username),
			})
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Get completed challenges
	completedChallengesRaw, err := repo.GetUserCompletedChallengesWithDetails(ctx.Request.Context(), username)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Get recent submissions
	recentSubmissionsRaw, err := repo.GetUserRecentSubmissions(ctx.Request.Context(), username, 20)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Get difficulty breakdown
	difficultyBreakdownRaw, err := repo.GetUserDifficultyBreakdown(ctx.Request.Context(), username)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Get language stats
	languageStatsRaw, err := repo.GetUserLanguageStats(ctx.Request.Context(), username)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Process and format all data
	profileData := processUserProfileData(userStats, completedChallengesRaw, recentSubmissionsRaw, difficultyBreakdownRaw, languageStatsRaw)

	if boost := ctx.Request.Header["Hx-Boosted"]; len(boost) == 0 {
		RenderHTML(ctx, http.StatusOK, "profile.html", profileData)
	} else {
		ctx.HTML(http.StatusOK, "user-profile", profileData)
	}
}

func processUserProfileData(
	userStats data.GetUserStatsRow,
	completedChallengesRaw []data.GetUserCompletedChallengesWithDetailsRow,
	recentSubmissionsRaw []data.GetUserRecentSubmissionsRow,
	difficultyBreakdownRaw []data.GetUserDifficultyBreakdownRow,
	languageStatsRaw []data.GetUserLanguageStatsRow,
) UserProfileData {

	// Calculate success rate
	successRate := "0.0%"
	if userStats.TotalSolutions > 0 {
		rate := (float64(userStats.AcceptedSolutions) / float64(userStats.TotalSolutions)) * 100
		successRate = fmt.Sprintf("%.1f%%", rate)
	}

	// Calculate days active
	daysActive := (time.Now().Unix() - userStats.CreatedAt) / 86400

	// Format member since date
	memberSince := time.Unix(userStats.CreatedAt, 0).Format("Jan 2, 2006")

	// Most used language
	mostUsedLanguage := "None"
	if len(languageStatsRaw) > 0 {
		mostUsedLanguage = languageStatsRaw[0].Language
	}

	// Process difficulty stats
	difficultyStats := make([]DifficultyStats, 0)
	difficultyLabels := map[int64]string{
		0: "Untagged",
		1: "Easy",
		2: "Medium",
		3: "Hard",
	}
	difficultyClasses := map[int64]string{
		0: "difficulty",
		1: "difficulty easy",
		2: "difficulty medium",
		3: "difficulty hard",
	}

	for _, d := range difficultyBreakdownRaw {
		label := difficultyLabels[d.Difficulty]
		cssClass := difficultyClasses[d.Difficulty]
		difficultyStats = append(difficultyStats, DifficultyStats{
			Label:    label,
			Count:    d.CompletedCount,
			CssClass: cssClass,
		})
	}

	// Process language stats
	languageStats := make([]LanguageStat, 0)
	for _, lang := range languageStatsRaw {
		acceptanceRate := float64(0)
		if lang.SubmissionCount > 0 {
			acceptanceRate = (float64(lang.AcceptedCount) / float64(lang.SubmissionCount)) * 100
		}
		languageStats = append(languageStats, LanguageStat{
			Language:        lang.Language,
			SubmissionCount: lang.SubmissionCount,
			AcceptedCount:   lang.AcceptedCount,
			AcceptanceRate:  acceptanceRate,
		})
	}

	// Process completed challenges
	completedChallenges := make([]CompletedChallenge, 0)
	for _, c := range completedChallengesRaw {
		diffLabel := difficultyLabels[c.Difficulty]
		diffClass := difficultyClasses[c.Difficulty]
		completedAt := formatRelativeTime(c.CompletedAt)

		completedChallenges = append(completedChallenges, CompletedChallenge{
			ChallengeID:       c.ChallengeID,
			Title:             c.Title,
			Difficulty:        c.Difficulty,
			DifficultyLabel:   diffLabel,
			DifficultyClass:   diffClass,
			Language:          c.Language,
			CompletedAt:       completedAt,
			RuntimeInfo:       c.RuntimeInfo.String,
			SolutionCreatedAt: formatRelativeTime(c.SolutionCreatedAt),
		})
	}

	// Process recent submissions
	recentSubmissions := make([]RecentSubmission, 0)
	for _, s := range recentSubmissionsRaw {
		statusClass := getStatusBadgeClass(s.Status)
		submittedAt := formatRelativeTime(s.CreatedAt)

		recentSubmissions = append(recentSubmissions, RecentSubmission{
			ID:               s.ID,
			ChallengeID:      s.ChallengeID,
			ChallengeTitle:   s.ChallengeTitle,
			Language:         s.Language,
			Status:           s.Status,
			StatusBadgeClass: statusClass,
			RuntimeInfo:      s.RuntimeInfo.String,
			SubmittedAt:      submittedAt,
		})
	}

	return UserProfileData{
		Username:               userStats.Username,
		IsAdmin:                userStats.IsAdmin == 1,
		MemberSince:            memberSince,
		DaysActive:             daysActive,
		ChallengesCompleted:    userStats.ChallengesCompleted,
		TotalSolutions:         userStats.TotalSolutions,
		AcceptedSolutions:      userStats.AcceptedSolutions,
		WrongAnswers:           userStats.WrongAnswers,
		RuntimeErrors:          userStats.RuntimeErrors,
		SuccessRate:            successRate,
		MostUsedLanguage:       mostUsedLanguage,
		DifficultyStats:        difficultyStats,
		LanguageStats:          languageStats,
		CompletedChallenges:    completedChallenges,
		RecentSubmissions:      recentSubmissions,
		HasCompletedChallenges: len(completedChallenges) > 0,
		HasRecentSubmissions:   len(recentSubmissions) > 0,
		HasLanguageStats:       len(languageStats) > 0,
	}
}

func formatRelativeTime(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	}
}

func getStatusBadgeClass(status string) string {
	switch status {
	case "accepted":
		return "status-badge status-accepted"
	case "wrong_answer":
		return "status-badge status-wrong"
	case "runtime_error":
		return "status-badge status-error"
	case "pending":
		return "status-badge status-pending"
	default:
		return "status-badge status-unknown"
	}
}

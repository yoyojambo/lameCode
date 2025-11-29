package app

import (
	"fmt"
	"lameCode/internal/platform/data"
	"lameCode/internal/web/ui"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func LoadUserProfileHandlers(r *gin.RouterGroup) {
	r.GET("/user/:username", enableHtmxCache, userProfilePageFunc)
}

func userProfilePageFunc(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("username is required"))
		return
	}

	repo := data.Repository()
	reqCtx := ctx.Request.Context()

	// Fetch all data directly
	userStats, err := repo.GetUserStats(reqCtx, username)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			// TODO: render a proper 404 templ page
			ctx.String(http.StatusNotFound, "User '%s' not found", username)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	completed, err := repo.GetUserCompletedChallengesWithDetails(reqCtx, username)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	submissions, err := repo.GetUserRecentSubmissions(reqCtx, username, 20)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	difficulty, err := repo.GetUserDifficultyBreakdown(reqCtx, username)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	languages, err := repo.GetUserLanguageStats(reqCtx, username)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Build the view directly - no processing needed
	profile := ui.ProfileView{
		User:                userStats,
		CompletedChallenges: completed,
		RecentSubmissions:   submissions,
		DifficultyBreakdown: difficulty,
		LanguageStats:       languages,
	}

	user := extractUserData(ctx)

	if ctx.GetHeader("HX-Boosted") != "" {
		RenderTemplOK(ctx, ui.ProfileContent(profile))
	} else {
		RenderTemplOK(ctx, ui.ProfilePage(user, profile))
	}
}

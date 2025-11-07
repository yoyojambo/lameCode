package app

import (
	"lameCode/internal/platform/data"

	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func LoadAdminHandlers(r *gin.RouterGroup) {
	// Admin dashboard
	r.GET("/admin", enableHtmxCache, adminDashboardFunc)

	// Problems management
	r.GET("/admin/problems", enableHtmxCache, adminProblemsListFunc)
	r.GET("/admin/problems/new", enableHtmxCache, adminCreateProblemPageFunc)
	r.POST("/admin/problems/new", adminCreateProblemFunc)
	r.GET("/admin/problems/:id/edit", enableHtmxCache, adminEditProblemPageFunc)
	r.POST("/admin/problems/:id/edit", adminUpdateProblemFunc)
	r.DELETE("/admin/problems/:id", adminDeleteProblemFunc)

	// Test case management
	r.GET("/admin/problems/:id/tests", enableHtmxCache, adminManageTestsPageFunc)
	r.POST("/admin/problems/:id/tests", adminCreateTestFunc)
	r.PUT("/admin/problems/:id/tests/:testId", adminUpdateTestFunc)
	r.DELETE("/admin/problems/:id/tests/:testId", adminDeleteTestFunc)
}

type AdminDashboardData struct {
	Stats data.GetAdminStatsRow
}

type AdminProblemRow struct {
	ID              int64
	Title           string
	Description     string
	TruncatedDesc   string
	Difficulty      int64
	DifficultyLabel string
	DifficultyClass string
	TestCount       int64
	SolverCount     int64
	SubmissionCount int64
	CreatedAt       int64
	FormattedDate   string
}

type AdminProblemsData struct {
	Problems    []AdminProblemRow
	HasProblems bool
}

type AdminTestCaseRow struct {
	ID              int64
	InputData       string
	ExpectedOutput  string
	TruncatedInput  string
	TruncatedOutput string
	TestNumber      int
}

type AdminProblemFormData struct {
	Problem         *data.Challenge
	Tests           []AdminTestCaseRow
	IsEdit          bool
	TestCount       int
	HasTests        bool
	DifficultyLabel string
}

func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

func getDifficultyLabel(d int64) string {
	switch d {
	case 1:
		return "Easy"
	case 2:
		return "Medium"
	case 3:
		return "Hard"
	default:
		return "Untagged"
	}
}

func getDifficultyClass(d int64) string {
	switch d {
	case 1:
		return "easy"
	case 2:
		return "medium"
	case 3:
		return "hard"
	default:
		return ""
	}
}

func formatUnixDate(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("Jan 2, 2006")
}

func transformProblemsForDisplay(problems []data.GetProblemsForAdminRow) []AdminProblemRow {
	var result []AdminProblemRow
	for _, p := range problems {
		result = append(result, AdminProblemRow{
			ID:              p.ID,
			Title:           p.Title,
			Description:     p.Description,
			TruncatedDesc:   truncateString(p.Description, 100),
			Difficulty:      p.Difficulty,
			DifficultyLabel: getDifficultyLabel(p.Difficulty),
			DifficultyClass: getDifficultyClass(p.Difficulty),
			TestCount:       p.TestCount,
			SolverCount:     p.SolverCount,
			SubmissionCount: p.SubmissionCount,
			CreatedAt:       p.CreatedAt,
			FormattedDate:   formatUnixDate(p.CreatedAt),
		})
	}
	return result
}

func adminDashboardFunc(ctx *gin.Context) {
	repo := data.Repository()

	stats, err := repo.GetAdminStats(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	dashboardData := AdminDashboardData{
		Stats: stats,
	}

	if boost := ctx.Request.Header["Hx-Boosted"]; len(boost) == 0 {
		RenderHTML(ctx, http.StatusOK, "admin_dashboard.html", dashboardData)
	} else {
		ctx.HTML(http.StatusOK, "admin-dashboard", dashboardData)
	}
}

func adminProblemsListFunc(ctx *gin.Context) {
	repo := data.Repository()

	// Just get all problems, no filtering.
	problems, err := repo.GetProblemsForAdmin(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	transformedProblems := transformProblemsForDisplay(problems)

	problemsData := AdminProblemsData{
		Problems:    transformedProblems,
		HasProblems: len(transformedProblems) > 0,
	}

	if boost := ctx.Request.Header["Hx-Boosted"]; len(boost) == 0 {
		RenderHTML(ctx, http.StatusOK, "admin_problems.html", problemsData)
	} else {
		ctx.HTML(http.StatusOK, "admin-problems", problemsData)
	}
}

func adminCreateProblemPageFunc(ctx *gin.Context) {
	formData := AdminProblemFormData{
		Problem:   nil,
		Tests:     []AdminTestCaseRow{},
		IsEdit:    false,
		TestCount: 0,
		HasTests:  false,
	}

	if boost := ctx.Request.Header["Hx-Boosted"]; len(boost) == 0 {
		RenderHTML(ctx, http.StatusOK, "admin_problem_form.html", formData)
	} else {
		ctx.HTML(http.StatusOK, "admin-problem-form", formData)
	}
}

func adminCreateProblemFunc(ctx *gin.Context) {
	var req struct {
		Title       string `form:"title" binding:"required,min=3,max=200"`
		Description string `form:"description" binding:"required,min=10"`
		Difficulty  int64  `form:"difficulty" binding:"required,min=0,max=3"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.HTML(http.StatusOK, "form-error", gin.H{
			"message": "Please fill in all required fields correctly",
		})
		return
	}

	repo := data.Repository()

	challengeID, err := repo.NewChallenge(ctx.Request.Context(),
		req.Title, req.Description, req.Difficulty)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Created new challenge with ID %d\n", challengeID)
	ctx.Header("HX-Redirect", fmt.Sprintf("/admin/problems/%d/edit", challengeID))
}

func adminEditProblemPageFunc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid problem ID"))
		return
	}

	repo := data.Repository()

	problem, err := repo.GetChallengeWithTests(ctx.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			ctx.HTML(http.StatusNotFound, "error.html", gin.H{
				"message": "Problem not found",
			})
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	tests, err := repo.GetTestsForChallenge(ctx.Request.Context(), id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Transform tests for display
	var transformedTests []AdminTestCaseRow
	for i, test := range tests {
		transformedTests = append(transformedTests, AdminTestCaseRow{
			ID:              test.ID,
			InputData:       test.InputData,
			ExpectedOutput:  test.ExpectedOutput,
			TruncatedInput:  truncateString(test.InputData, 50),
			TruncatedOutput: truncateString(test.ExpectedOutput, 50),
			TestNumber:      i + 1,
		})
	}

	formData := AdminProblemFormData{
		Problem:         &problem,
		Tests:           transformedTests,
		IsEdit:          true,
		TestCount:       len(transformedTests),
		HasTests:        len(transformedTests) > 0,
		DifficultyLabel: getDifficultyLabel(problem.Difficulty),
	}

	if boost := ctx.Request.Header["Hx-Boosted"]; len(boost) == 0 {
		RenderHTML(ctx, http.StatusOK, "admin_problem_form.html", formData)
	} else {
		ctx.HTML(http.StatusOK, "admin-problem-form", formData)
	}
}

func adminUpdateProblemFunc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid problem ID"))
		return
	}

	var req struct {
		Title       string `form:"title" binding:"required,min=3,max=200"`
		Description string `form:"description" binding:"required,min=10"`
		Difficulty  int64  `form:"difficulty" binding:"required,min=0,max=3"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.HTML(http.StatusOK, "form-error", gin.H{
			"message": "Please fill in all required fields correctly",
		})
		return
	}

	repo := data.Repository()

	_, err = repo.UpdateChallenge(ctx.Request.Context(),
		req.Title, req.Description, req.Difficulty, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Updated challenge with ID %d\n", id)
	ctx.HTML(http.StatusOK, "success-message", gin.H{
		"message": "Problem updated successfully!",
	})
}

func adminDeleteProblemFunc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid problem ID"))
		return
	}

	repo := data.Repository()

	err = repo.DeleteChallenge(ctx.Request.Context(), id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Deleted challenge with ID %d\n", id)
	ctx.Header("HX-Redirect", "/admin/problems")
}

func adminManageTestsPageFunc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid problem ID"))
		return
	}

	repo := data.Repository()

	problem, err := repo.GetChallengeWithTests(ctx.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			ctx.HTML(http.StatusNotFound, "error.html", gin.H{
				"message": "Problem not found",
			})
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	tests, err := repo.GetTestsForChallenge(ctx.Request.Context(), id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Transform tests for display
	var transformedTests []AdminTestCaseRow
	for i, test := range tests {
		transformedTests = append(transformedTests, AdminTestCaseRow{
			ID:             test.ID,
			InputData:      test.InputData,
			ExpectedOutput: test.ExpectedOutput,
			TestNumber:     i + 1,
		})
	}

	formData := AdminProblemFormData{
		Problem:   &problem,
		Tests:     transformedTests,
		IsEdit:    true,
		TestCount: len(transformedTests),
		HasTests:  len(transformedTests) > 0,
	}

	if boost := ctx.Request.Header["Hx-Boosted"]; len(boost) == 0 {
		RenderHTML(ctx, http.StatusOK, "admin_tests.html", formData)
	} else {
		ctx.HTML(http.StatusOK, "admin-tests", formData)
	}
}

func adminCreateTestFunc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid problem ID"))
		return
	}

	var req struct {
		Input  string `form:"input" binding:"required"`
		Output string `form:"output" binding:"required"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.HTML(http.StatusOK, "form-error", gin.H{
			"message": "Both input and output are required",
		})
		return
	}

	repo := data.Repository()

	testID, err := repo.NewChallengeTest(ctx.Request.Context(), id, req.Input, req.Output)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Created new test case with ID %d for challenge %d\n", testID, id)
	ctx.Header("HX-Refresh", "true")
}

func adminUpdateTestFunc(ctx *gin.Context) {
	testIdStr := ctx.Param("testId")
	testId, err := strconv.ParseInt(testIdStr, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid test ID"))
		return
	}

	var req struct {
		Input  string `form:"input" binding:"required"`
		Output string `form:"output" binding:"required"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.HTML(http.StatusOK, "form-error", gin.H{
			"message": "Both input and output are required",
		})
		return
	}

	repo := data.Repository()

	_, err = repo.UpdateChallengeTest(ctx.Request.Context(), req.Input, req.Output, testId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Updated test case with ID %d\n", testId)
	ctx.HTML(http.StatusOK, "success-message", gin.H{
		"message": "Test case updated successfully!",
	})
}

func adminDeleteTestFunc(ctx *gin.Context) {
	testIdStr := ctx.Param("testId")
	testId, err := strconv.ParseInt(testIdStr, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid test ID"))
		return
	}

	repo := data.Repository()

	err = repo.DeleteChallengeTest(ctx.Request.Context(), testId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Deleted test case with ID %d\n", testId)
	ctx.Header("HX-Refresh", "true")
}

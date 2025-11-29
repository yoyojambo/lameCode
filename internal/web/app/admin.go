package app

import (
	"fmt"
	"lameCode/internal/platform/data"
	"lameCode/internal/web/ui"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func LoadAdminHandlers(r *gin.RouterGroup) {
	r.GET("/admin", enableHtmxCache, adminDashboardFunc)
	r.GET("/admin/problems", enableHtmxCache, adminProblemsListFunc)
	r.GET("/admin/problems/new", enableHtmxCache, adminCreateProblemPageFunc)
	r.POST("/admin/problems/new", adminCreateProblemFunc)
	r.GET("/admin/problems/:id/edit", enableHtmxCache, adminEditProblemPageFunc)
	r.POST("/admin/problems/:id/edit", adminUpdateProblemFunc)
	r.DELETE("/admin/problems/:id", adminDeleteProblemFunc)
	r.GET("/admin/problems/:id/tests", enableHtmxCache, adminManageTestsPageFunc)
	r.POST("/admin/problems/:id/tests", adminCreateTestFunc)
	r.PUT("/admin/problems/:id/tests/:testId", adminUpdateTestFunc)
	r.DELETE("/admin/problems/:id/tests/:testId", adminDeleteTestFunc)
}

func adminDashboardFunc(ctx *gin.Context) {
	repo := data.Repository()
	reqCtx := ctx.Request.Context()

	stats, err := repo.GetAdminStats(reqCtx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	view := ui.AdminDashboardView{GetAdminStatsRow: stats}
	user := extractUserData(ctx)

	if isHtmxBoosted(ctx) {
		ui.AdminDashboardContent(view).Render(reqCtx, ctx.Writer)
	} else {
		ui.AdminDashboardPage(user, view).Render(reqCtx, ctx.Writer)
	}
}

func adminProblemsListFunc(ctx *gin.Context) {
	repo := data.Repository()
	reqCtx := ctx.Request.Context()

	problems, err := repo.GetProblemsForAdmin(reqCtx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	view := ui.AdminProblemsView{Problems: problems}
	user := extractUserData(ctx)

	if isHtmxBoosted(ctx) {
		ui.AdminProblemsContent(view).Render(reqCtx, ctx.Writer)
	} else {
		ui.AdminProblemsPage(user, view).Render(reqCtx, ctx.Writer)
	}
}

func adminCreateProblemPageFunc(ctx *gin.Context) {
	view := ui.AdminProblemFormView{
		Problem: nil,
		Tests:   []data.ChallengeTest{},
	}
	user := extractUserData(ctx)

	if isHtmxBoosted(ctx) {
		ui.AdminProblemFormContent(view).Render(ctx.Request.Context(), ctx.Writer)
	} else {
		ui.AdminProblemFormPage(user, view).Render(ctx.Request.Context(), ctx.Writer)
	}
}

func adminCreateProblemFunc(ctx *gin.Context) {
	var req struct {
		Title       string `form:"title" binding:"required,min=3,max=200"`
		Description string `form:"description" binding:"required,min=10"`
		Difficulty  int64  `form:"difficulty" binding:"required,min=0,max=3"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ui.FormErrorMessage("Please fill in all required fields correctly").Render(ctx.Request.Context(), ctx.Writer)
		return
	}

	repo := data.Repository()
	challengeID, err := repo.NewChallenge(ctx.Request.Context(), req.Title, req.Description, req.Difficulty)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Created new challenge with ID %d\n", challengeID)
	ctx.Header("HX-Redirect", fmt.Sprintf("/admin/problems/%d/edit", challengeID))
}

func adminEditProblemPageFunc(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid problem ID"))
		return
	}

	repo := data.Repository()
	reqCtx := ctx.Request.Context()

	problem, err := repo.GetChallengeWithTests(reqCtx, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			ctx.String(http.StatusNotFound, "Problem not found")
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	tests, err := repo.GetTestsForChallenge(reqCtx, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	view := ui.AdminProblemFormView{
		Problem: &problem,
		Tests:   tests,
	}
	user := extractUserData(ctx)

	if isHtmxBoosted(ctx) {
		ui.AdminProblemFormContent(view).Render(reqCtx, ctx.Writer)
	} else {
		ui.AdminProblemFormPage(user, view).Render(reqCtx, ctx.Writer)
	}
}

func adminUpdateProblemFunc(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
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
		ui.FormErrorMessage("Please fill in all required fields correctly").Render(ctx.Request.Context(), ctx.Writer)
		return
	}

	repo := data.Repository()
	_, err = repo.UpdateChallenge(ctx.Request.Context(), req.Title, req.Description, req.Difficulty, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Updated challenge with ID %d\n", id)
	ui.FormSuccessMessage("Problem updated successfully!").Render(ctx.Request.Context(), ctx.Writer)
}

func adminDeleteProblemFunc(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid problem ID"))
		return
	}

	repo := data.Repository()
	if err := repo.DeleteChallenge(ctx.Request.Context(), id); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Deleted challenge with ID %d\n", id)
	//ctx.Header("HX-Redirect", "/admin/problems")
}

func adminManageTestsPageFunc(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid problem ID"))
		return
	}

	repo := data.Repository()
	reqCtx := ctx.Request.Context()

	problem, err := repo.GetChallengeWithTests(reqCtx, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			ctx.String(http.StatusNotFound, "Problem not found")
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	tests, err := repo.GetTestsForChallenge(reqCtx, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	view := ui.AdminTestsView{
		Problem: problem,
		Tests:   tests,
	}
	user := extractUserData(ctx)

	if isHtmxBoosted(ctx) {
		ui.AdminTestsContent(view).Render(reqCtx, ctx.Writer)
	} else {
		ui.AdminTestsPage(user, view).Render(reqCtx, ctx.Writer)
	}
}

func adminCreateTestFunc(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid problem ID"))
		return
	}

	var req struct {
		Input  string `form:"input" binding:"required"`
		Output string `form:"output" binding:"required"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ui.FormErrorMessage("Both input and output are required").Render(ctx.Request.Context(), ctx.Writer)
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
	testId, err := strconv.ParseInt(ctx.Param("testId"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid test ID"))
		return
	}

	var req struct {
		Input  string `form:"input" binding:"required"`
		Output string `form:"output" binding:"required"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ui.FormErrorMessage("Both input and output are required").Render(ctx.Request.Context(), ctx.Writer)
		return
	}

	repo := data.Repository()
	_, err = repo.UpdateChallengeTest(ctx.Request.Context(), req.Input, req.Output, testId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Updated test case with ID %d\n", testId)
	ctx.Header("HX-Refresh", "true")
}

func adminDeleteTestFunc(ctx *gin.Context) {
	testId, err := strconv.ParseInt(ctx.Param("testId"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid test ID"))
		return
	}

	repo := data.Repository()
	if err := repo.DeleteChallengeTest(ctx.Request.Context(), testId); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Deleted test case with ID %d\n", testId)
	ctx.Header("HX-Refresh", "true")
}


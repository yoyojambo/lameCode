package app

import (
	"context"
	"fmt"
	"lameCode/internal/platform/data"
	"lameCode/internal/platform/judge"
	"lameCode/internal/web/ui"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Submission struct {
	Language string `form:"language" binding:"required"`
	Code     string `form:"code" binding:"required"`
}

func LoadJudgeHandlers(r *gin.RouterGroup) {
	g := r.Group("/judge")
	g.POST("/test/:id", testSubmission)
	g.POST("/submit/:id", printSubmission) // TODO: Break out tests behaviour and actually save submissions with submit
}

// printSubmission is just for testing the frontend
// Response is swapped by HTMX
func printSubmission(ctx *gin.Context) {
	var submission Submission
	if err := ctx.ShouldBind(&submission); err != nil {
		ctx.AbortWithError(http.StatusBadRequest,
			fmt.Errorf("could not bind context to Submission object: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

func testSubmission(ctx *gin.Context) {
	challengeId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest,
			fmt.Errorf("error parsing challenge id in /test: %w", err))
		
		return
	}

	var submission Submission
	if err := ctx.ShouldBind(&submission); err != nil {
		ctx.AbortWithError(http.StatusBadRequest,
			fmt.Errorf("could not bind code submission: %w", err))
		return
	}

	q := data.Repository()
	testCtx := context.Background()
	tests, err := q.GetTestsForChallenge(testCtx, challengeId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError,
			fmt.Errorf("error getting tests for challenge in test: %w", err))
		
		return
	}

	results, err := judge.RunMultipleTests(submission.Code, submission.Language, tests)
	// Check if this is compiler error (bad code) or internal error
	if err != nil {
		// Compilation error case (user error message)
		if strings.HasPrefix(err.Error(), "Error compiling") {
			nLine := strings.Index(err.Error(), "\n")
			errmsg := err.Error()[nLine+1:]
			RenderTemplOK(ctx, ui.CompilerMessage(errmsg))
			return
		}
		// Error in any other phase
		ctx.AbortWithError(http.StatusInternalServerError,
			fmt.Errorf("error running tests: %w", err))
		return
	}
	
	RenderTemplOK(ctx, ui.ResultTable(results))
}

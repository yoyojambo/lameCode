package app

import (
	"context"
	"lameCode/platform/data"
	"lameCode/platform/judge"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Submission struct {
	Language string `form:"language" binding:"required"`
	Code     string `form:"code" binding:"required"`
}

func LoadJudgeHandlers(r *gin.Engine) {
	g := r.Group("/judge")
	g.POST("/test/:id", testSubmission)
	g.POST("/submit/:id", printSubmission)
}

// printSubmission is just for testing the frontend
// Response is swapped by HTMX
func printSubmission(ctx *gin.Context) {
	var submission Submission
	if err := ctx.ShouldBind(&submission); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		l.Println(err)
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

func testSubmission(ctx *gin.Context) {
	challengeId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		l.Println("Error parsing challenge id in /test:", err)
	}

	var submission Submission
	if err := ctx.ShouldBind(&submission); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		l.Println(err)
		return
	}

	q := data.Repository()
	testCtx := context.Background()
	tests, err := q.GetTestsForChallenge(testCtx, challengeId)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		l.Println("Error getting tests for challenge in test:", err)
	}

	results, err := judge.RunMultipleTests(submission.Code, submission.Language, tests)
	// Check if this is compiler error (bad code) or internal error
	if err != nil {
		// Compilation error case (user error message)
		if strings.HasPrefix(err.Error(), "Error compiling") {
			nLine := strings.Index(err.Error(), "\n")
			errmsg := err.Error()[nLine+1:]
			ctx.HTML(http.StatusOK, "compiler-message", gin.H{"Message": errmsg})
			return
		}
		// Error in any other phase
		ctx.AbortWithStatus(http.StatusInternalServerError)
		l.Println("Error running tests for challenge in test:", err)
	}
	
	ctx.HTML(http.StatusOK, "result-table", results)
}

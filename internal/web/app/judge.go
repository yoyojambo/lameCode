package app

import (
	"database/sql"
	"errors"
	"fmt"
	"lameCode/internal/platform/data"
	"lameCode/internal/platform/judge"
	"lameCode/internal/platform/session"
	"lameCode/internal/web/ui"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func LoadJudgeHandlers(r *gin.RouterGroup) {
	g := r.Group("/judge")
	g.POST("/test/:id", evalSolutionHandlers(true))
	g.POST("/submit/:id", session.MandatoryAuthRoute("/login"), evalSolutionHandlers(false)) // TODO: Break out tests behaviour and actually save submissions with submit
}

// printSubmission is just for testing the frontend
// Response is swapped by HTMX
func printSubmission(ctx *gin.Context) {
	var submission ui.Submission
	if err := ctx.ShouldBind(&submission); err != nil {
		ctx.AbortWithError(http.StatusBadRequest,
			fmt.Errorf("could not bind context to Submission object: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

func evalSolutionHandlers(just_test bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		challengeId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest,
				fmt.Errorf("error parsing challenge id in /test: %w", err))

			return
		}

		var submission ui.Submission
		if err := ctx.ShouldBind(&submission); err != nil {
			ctx.AbortWithError(http.StatusBadRequest,
				fmt.Errorf("could not bind code submission: %w", err))
			return
		}

		q := data.Repository()
		tests, err := q.GetTestsForChallenge(ctx.Request.Context(), challengeId)
		if err != nil {
			// Apparently does not work? maybe because :many sqlc hint?
			if errors.Is(err, sql.ErrNoRows) {
				ctx.String(http.StatusOK, "No tests for challenge...")
			} else {
				ctx.AbortWithError(http.StatusInternalServerError,
					fmt.Errorf("error getting tests for challenge in test: %w", err))
			}
			return
		}

		if len(tests) == 0 {
			ctx.String(http.StatusOK, "<b>No tests for challenge...<b>")
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

		// if a submission, not just test, save to DB
		// assumes the user is logged in (check with middleware)
		if !just_test {
			user := extractUserData(ctx)
			status := "accepted"

			count := 0
			for _, v := range results {
				if v.Pass {
					count++
				}
			}

			if count != len(results) {
				status = "wrong_answer"
			}

			sol_id, err := q.NewSolutionByUsername(ctx.Request.Context(),
				data.NewSolutionByUsernameParams{
					Username:    user.Username,
					ChallengeID: challengeId,
					Code:        submission.Code,
					Language:    submission.Language,
					Status:      status,
					RuntimeInfo: sql.NullString{
						String: "",
						Valid:  true,
					},
				})

			if err != nil {
				RenderTemplOK(ctx, ui.CompilerMessage("Internal pipeline error!"))
				l.Printf("Error saving submission: %v", err)
				return
			}

			l.Printf("New submission (%d) received for problem '%d' by %s",
				sol_id, challengeId, user.Username)
		}

		RenderTemplOK(ctx, ui.ResultTable(submission, results))
	}
}

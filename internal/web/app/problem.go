package app

import (
	"database/sql"
	"errors"
	"lameCode/internal/platform/config"
	"lameCode/internal/platform/data"
	"lameCode/internal/platform/judge"
	"lameCode/internal/web/ui"

	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func LoadProblemHandlers(r *gin.RouterGroup) {
	r.GET("/", problemSetFunc)
	r.GET("/problemlist", enableHtmxCache, problemsSetPageFunc)
	r.GET("/problem/:id", enableHtmxCache, problemFunc)
}

// If the Markdown in Description field begins with "# " indicating a h1
// title, it will the text in that line the Title field.
func tryBetterTitle(challenge *data.Challenge) {
	if len(challenge.Description) < 2 {
		if config.Debug() {
			l.Printf("Attempted better title on \n%+v:\nDescription too short!", *challenge)
		}
		return
	}

	challenge.Description = strings.TrimSpace(challenge.Description)

	// Assume it is starting with a title, so assign it
	if challenge.Description[0:2] == "# " {
		tIdx := strings.Index(challenge.Description, "\n")
		challenge.Title = challenge.Description[2:tIdx] // Title is the part after "# " in Markdown
		challenge.Description = challenge.Description[tIdx+1:]
	}
}

func problemFunc(ctx *gin.Context) {
	// Let it check compilers in the background, so it can cache it
	compilers := judge.LanguageOptions()
	if config.Debug() {
		go func(compilers []judge.LanguageOption) {
			for _, c := range compilers {
				l.Printf("%+v\n", c)
			}
		}(compilers)
	}

	problemId_str := ctx.Param("id")
	problemId, err := strconv.ParseInt(problemId_str, 10, 64)
	if err != nil {
		ctx.AbortWithError(500, err)
	}

	p, err := data.Repository().GetChallenge(ctx, problemId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		ctx.AbortWithError(500, err)
		return
	}

	// TODO: Make 404 not found page.

	// wrapped the literal in () because otherwise it thinks {} is the
	// beginning of an expression
	if p == (data.Challenge{}) { // compare to 0-value in case of not found
		ctx.AbortWithStatus(404)
		return
	}

	tryBetterTitle(&p)

	if ctx.GetHeader("HX-Request") == "true" {
		RenderTemplOK(ctx, ui.ProblemEditor(p, judge.LanguageOptions()))
	} else {
		RenderTemplOK(ctx, ui.ProblemPage(extractUserData(ctx), p, judge.LanguageOptions()))
	}
}

func problemSetFunc(ctx *gin.Context) {
	// Generate data
	pageData, err := getPageData(ctx, 1)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if ctx.GetHeader("HX-Request") == "true" {
		ctx.HTML(http.StatusOK, "problemTable", pageData)
	} else {
		RenderHTML(ctx, http.StatusOK, "problems.html", pageData)
	}
}

func problemsSetPageFunc(ctx *gin.Context) {
	pageStr := ctx.Query("page")

	page, err := strconv.ParseInt(pageStr, 10, 64) // Straigt to int64
	if err != nil || page < 1 {                    // fails over to first page
		page = 1
	}

	pageData, err := getPageData(ctx, page)
	if err != nil {
		l.Printf("Error generating page of problems: %v", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Render the partial template for HTMX.
	if len(pageData.Challenges) == 0 {
		l.Println("Not showing any challenges")
	}

	ctx.HTML(http.StatusOK, "challengeList", pageData)
}

// Information for a list of challenges, from a paged request
type ChallengePage struct {
	Challenges  []data.Challenge `json:"challenges"`
	HasPrev     bool             `json:"has_prev"`
	HasNext     bool             `json:"has_next"`
	PrevPage    int64            `json:"prev_page"`
	NextPage    int64            `json:"next_page"`
	CurrentPage int64            `json:"current_page"`
}

func getPageData(ctx *gin.Context, page int64) (ChallengePage, error) {
	const pageSize = 10
	offset := (page - 1) * pageSize

	// Query the paginated challenges.
	challenges_data, err := data.Repository().GetChallengesPaginated(ctx, pageSize+1, offset)

	if err != nil {
		l.Printf("error fetching paginated challenges: %v", err)
		return ChallengePage{}, err
	}
	if len(challenges_data) == 0 {
		l.Println("No challenges found...")
	}

	// Get better titles if available
	for i := range challenges_data {
		tryBetterTitle(&challenges_data[i])
	}

	// Determine if previous and next pages exist.
	hasPrev := page > 1
	// Assume there is a next page unless less than pagesize challenges returned
	hasNext := len(challenges_data) > pageSize

	// Build the page structure.
	return ChallengePage{
		Challenges:  challenges_data,
		HasPrev:     hasPrev,
		HasNext:     hasNext,
		PrevPage:    page - 1,
		NextPage:    page + 1,
		CurrentPage: page,
	}, nil
}

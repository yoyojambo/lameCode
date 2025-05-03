package app

import (
	"html/template"
	"lameCode/platform/config"
	"lameCode/platform/data"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func LoadProblemHandlers(r *gin.Engine) {
	r.GET("/", problemSetFunc)
	r.GET("/problemlist", problemsSetPageFunc)
	r.GET("/problem/:id", problemFunc)
}

// Local representation of a challenge
// User has only information necessary to display on pages
type User struct {
	LoggedIn bool
	Username string
	Avatar   string // or empty if none
}

func fromUser(user data.User) User {
	return User{
		LoggedIn: true,
		Username: user.Username,
		Avatar:   "",
	}
}

func fromUsername(ctx *gin.Context, username string) User {
	user, err := data.Repository().GetUserByName(ctx, username)
	if err != nil {
		panic(err)
	}

	return fromUser(user)
}

// Local representation of a challenge.
// Necessary so Gin renders the HTML description correctly.
type Challenge struct {
	Id          int64
	Title       string
	Difficulty  int64
	Description template.HTML
}

// Straight up from https://github.com/gomarkdown/markdown
// Thanks for the library.
func mdToHTML(md string) string {
	md_bs := []byte(md)
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md_bs)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}

func fromChallenge(challenge data.Challenge) Challenge {
	return Challenge{
		Id:    challenge.ID,
		Title: challenge.Title,
		// Essentially a cast :/
		Description: template.HTML(mdToHTML(challenge.Description)),
		Difficulty:  challenge.Difficulty,
	}
}

func problemFunc(ctx *gin.Context) {
	problemId_str := ctx.Param("id")
	problemId, err := strconv.Atoi(problemId_str)
	if err != nil {
		ctx.AbortWithError(500, err)
	}

	p, err := data.Repository().GetChallenge(ctx, int64(problemId))
	if err != nil {
		ctx.AbortWithError(500, err)
	}

	data := gin.H{
		"User": User{
			LoggedIn: false,
		},

		// fromChallenge creates an object with the unescaped Descrtiption
		"Problem": fromChallenge(p),
	}

	// Handle caching
	ctx.Header("Vary", "HX-Boosted")
	if !config.Debug() {
		ctx.Header("Cache-Control", "public, max-age=1800")
	} else {
		ctx.Header("Cache-Control", "max-age=600, must-revalidate")
	}

	tmpl := "problem.html"
	if ctx.GetHeader("HX-Request") == "true" {
		tmpl = "problem"
	}
	ctx.HTML(http.StatusOK, tmpl, data)
}

func problemSetFunc(ctx *gin.Context) {
	// Handle caching
	ctx.Header("Vary", "HX-Boosted")
	if !config.Debug() {
		ctx.Header("Cache-Control", "public, max-age=1800")
	} else {
		ctx.Header("Cache-Control", "max-age=600, must-revalidate")
	}

	// Generate data
	tmpl := "problems.html"
	if ctx.GetHeader("HX-Request") == "true" {
		tmpl = "problemTable"
	}

	pageData := getPageData(ctx, 1)

	ctx.HTML(http.StatusOK, tmpl, pageData)
}

func problemsSetPageFunc(ctx *gin.Context) {
	pageStr := ctx.Query("page")

	//page, err := strconv.Atoi(pageStr)
	page, err := strconv.ParseInt(pageStr, 10, 64) // Straigt to int64
	if err != nil || page < 1 {
		page = 1
	}

	pageData := getPageData(ctx, page)

	// Render the partial template for HTMX.
	ctx.HTML(http.StatusOK, "challengeList", pageData)
}

// Information for a list of challenges, from a paged request
type ChallengePage struct {
	Challenges  []data.GetChallengesPaginatedRow `json:"challenges"`
	HasPrev     bool                             `json:"has_prev"`
	HasNext     bool                             `json:"has_next"`
	PrevPage    int64                            `json:"prev_page"`
	NextPage    int64                            `json:"next_page"`
	CurrentPage int64                            `json:"current_page"`
}

func getPageData(ctx *gin.Context, page int64) ChallengePage {
	repo := data.Repository()

	const pageSize = 30
	offset := (page - 1) * pageSize

	// Query the paginated challenges.
	challenges_data, err := repo.GetChallengesPaginated(ctx, pageSize, offset)
	if err != nil {
		log.Printf("error fetching paginated challenges: %v", err)
		return ChallengePage{}
	}

	// Optionally, retrieve the total count to determine pagination links.
	countRow, err := repo.CountChallenges(ctx)
	if err != nil {
		log.Printf("error counting challenges: %v", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return ChallengePage{}
	}

	// Determine if previous and next pages exist.
	hasPrev := page > 1
	hasNext := (page * pageSize) < countRow

	// Build the page structure.
	return ChallengePage{
		Challenges:  challenges_data,
		HasPrev:     hasPrev,
		HasNext:     hasNext,
		PrevPage:    page - 1,
		NextPage:    page + 1,
		CurrentPage: page,
	}
}

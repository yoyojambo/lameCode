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

// Information for a list of challenges, from a paged request
type ChallengePage struct {
	Challenges  []data.Challenge `json:"challenges"`
	HasPrev     bool             `json:"has_prev"`
	HasNext     bool             `json:"has_next"`
	PrevPage    int64            `json:"prev_page"`
	NextPage    int64            `json:"next_page"`
	CurrentPage int64            `json:"current_page"`
}

// Local representation of a challenge.
// Necessary so Gin renders the HTML description correctly.
type Challenge struct {
	Title string
	Difficulty int64
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
		Title: challenge.Title,
		                           // Essentially a cast :/
		Description: template.HTML(mdToHTML(challenge.Description)),
		Difficulty: challenge.Difficulty,
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
	if !config.Debug() {
		ctx.Header("Cache-Control", "public, max-age=31536000")
	}
	ctx.HTML(http.StatusOK, "problem.html", data)
}

func problemSetFunc(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "problems.html", nil)
}

func problemsSetPageFunc(ctx *gin.Context) {
	repo := data.Repository()

	pageStr := ctx.Query("page")
	//page, err := strconv.Atoi(pageStr)
	page, err := strconv.ParseInt(pageStr, 10, 64) // Straigt to int64
	if err != nil || page < 1 {
		page = 1
	}

	const pageSize = 30
	offset := (page - 1) * pageSize

	// Query the paginated challenges.
	challenges, err := repo.GetChallengesPaginated(ctx, pageSize, offset)
	if config.Debug() {
		log.Printf("Problems:\n  PAGE: %d\n  OFFSET: %d\n", page, offset)
	}
	if err != nil {
		log.Printf("error fetching paginated challenges: %v", err)
		ctx.String(http.StatusInternalServerError, "Internal server error")
		return
	}

	// Optionally, retrieve the total count to determine pagination links.
	countRow, err := repo.CountChallenges(ctx)
	if err != nil {
		log.Printf("error counting challenges: %v", err)
		ctx.String(http.StatusInternalServerError, "Internal server error")
		return
	}

	// Determine if previous and next pages exist.
	hasPrev := page > 1
	hasNext := (page * pageSize) < countRow

	// Build the page structure.
	pageData := ChallengePage{
		Challenges:  challenges,
		HasPrev:     hasPrev,
		HasNext:     hasNext,
		PrevPage:    page - 1,
		NextPage:    page + 1,
		CurrentPage: page,
	}

	// Render the partial template for HTMX.
	ctx.HTML(http.StatusOK, "challengeList", pageData)
}

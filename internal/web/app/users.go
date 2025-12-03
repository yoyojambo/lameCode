package app

import (
	"errors"
	"fmt"
	"lameCode/internal/platform/config"
	"lameCode/internal/platform/data"
	"lameCode/internal/platform/session"
	"lameCode/internal/web/ui"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// The regex for username and password form validation
var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{5,32}$`)
	passwordRegex = regexp.MustCompile(`^[A-Za-z0-9!@#$%^&*()_+\-={}\[\]:;<>,.?~]{8,70}$`)
)

func LoadUserHandlers(r *gin.RouterGroup) {
	r.GET("/login", enableHtmxCache, loginPageFunc)

	r.POST("/login", loginUserFunc)
	r.POST("/register", registerUserFunc)

	r.GET("/logout", logoutUserFunc)
}

func setSessionCookieFromToken(ctx *gin.Context, tok string, maxAge int) {
	ctx.SetCookie(session.SessionCookieName, tok, maxAge, "/", ctx.Request.Host, true, true)
}

// Just unsets the cookie and does clean pull for the login page
func logoutUserFunc(ctx *gin.Context) {
	setSessionCookieFromToken(ctx, "", 1)
	ctx.Header("HX-Redirect", "/login")
}

// Login an Register pages
func loginPageFunc(ctx *gin.Context) {
	if ctx.GetHeader("Hx-Request") == "true" {
		RenderTemplOK(ctx, ui.LoginContent(ui.FlashMessage{}))
	} else {
		RenderTemplOK(ctx, ui.LoginPage(extractUserData(ctx), ui.FlashMessage{}))
	}
}

// Message should always be consistent, so as to not give away which case occurred
const passwordMsg = "User or password is incorrect"

func loginUserFunc(ctx *gin.Context) {
	var req struct {
		Username string `form:"username" binding:"required,min=5,max=32"`
		Password string `form:"password" binding:"required,min=8,max=70"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		l.Println("Failed login request. Could not bind login form: ", err)
		RenderTemplOK(ctx, ui.FlashMessagePartial(
			ui.FlashMessage{
				Type: "error", Message: "Username or password is not of valid lenght"}))
		return
	}

	// Check content validity
	if !usernameRegex.MatchString(req.Username) || !passwordRegex.MatchString(req.Password) {
		RenderTemplOK(ctx, ui.FlashMessagePartial(
			ui.FlashMessage{
				Type: "error", Message: passwordMsg}))
		return
	}

	repo := data.Repository()

	user, err := repo.GetUserByName(ctx.Request.Context(), req.Username)
	if err != nil {
		// If user not found
		if strings.Contains(err.Error(), "no rows") {
			RenderTemplOK(ctx, ui.FlashMessagePartial(
				ui.FlashMessage{Type: "error", Message: passwordMsg}))
		} else { // Anything else
			ctx.AbortWithError(http.StatusInternalServerError,
				fmt.Errorf("error getting user: %w", err))
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(req.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			RenderTemplOK(ctx, ui.FlashMessagePartial(
				ui.FlashMessage{Type: "error", Message: passwordMsg}))
		} else {
			l.Printf("Comparison of hash %s and %s failed: %v\n",
				user.PasswordHash, req.Password, err)
			ctx.AbortWithError(http.StatusInternalServerError,
				fmt.Errorf("error comparing password: %w", err))
		}
		return
	}

	// TODO: JWT creation and user session middleware
	access := "user"
	if user.IsAdmin == 1 {
		access = "admin"
	}

	newToken, err := session.CreateSignedJwtToken(user.Username, access)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError,
			fmt.Errorf("could not create and sign new token: %w", err))
	}

	l.Printf("Logged in as %s!\n", req.Username)
	ctx.Header("HX-Redirect", "/")
	setSessionCookieFromToken(ctx, newToken, 99999)
}

func registerUserFunc(ctx *gin.Context) {
	var req struct {
		Username     string `form:"username" binding:"required,alphanum,min=5,max=32"`
		Password     string `form:"password" binding:"required,min=8,max=70"`
		Confirmation string `form:"confirm_password" binding:"required,min=8,max=70"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		RenderTemplOK(ctx, ui.FlashMessagePartial(
			ui.FlashMessage{
				Type: "error", Message: "Username or password is not of valid lenght"}))
		return
	}

	// Flash message
	if req.Password != req.Confirmation {
			RenderTemplOK(ctx, ui.FlashMessagePartial(
				ui.FlashMessage{Type: "error", Message: "Passwords don't match"}))

		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 0)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError,
			fmt.Errorf("Password hashing failed: %w", err))
		return
	} else if config.Debug() {
		l.Printf("Registering new user %s\n", req.Username)
	}

	repo := data.Repository()

	userID, err := repo.NewUser(ctx.Request.Context(),
		req.Username, hash)
	if err != nil {
		// Handle already existing user
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			l.Printf("Cannot register an existing user (%s)\n", req.Username)

			// Flash message response
			RenderTemplOK(ctx, ui.FlashMessagePartial(
				ui.FlashMessage{
					Type: "error",
					Message: fmt.Sprintf("User %s already exists...", req.Username)}))

		} else {
			ctx.AbortWithError(http.StatusInternalServerError,
				fmt.Errorf("could not create new user in DB: %w", err))
		}
		return
	}

	l.Printf("Created new user %s with ID %d\n", req.Username, userID)

	// Can only create a normal user, of course
	// TODO: Maybe be able to add a new admin if already an admin??
	newToken, err := session.CreateSignedJwtToken(req.Username, "user")
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError,
			fmt.Errorf("could not create and sign new token: %w", err))
	}

	// Redirect to homepage
	ctx.Header("HX-Redirect", "/")
	setSessionCookieFromToken(ctx, newToken, 99999)
}

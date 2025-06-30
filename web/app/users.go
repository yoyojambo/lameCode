package app

import (
	"errors"
	"fmt"
	"lameCode/platform/data"
	"lameCode/platform/session"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
	if boost := ctx.Request.Header["Hx-Boosted"]; len(boost) == 0 {
		ctx.HTML(http.StatusOK, "login.html", gin.H{})
	} else {
		ctx.HTML(http.StatusOK, "login", gin.H{})
	}
}

func loginUserFunc(ctx *gin.Context) {
	var req struct {
		Username     string `form:"username" binding:"required,alphanum,min=5,max=32"`
		Password     string `form:"password" binding:"required,min=8,max=70"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("could not bind login form: %w", err))
		return
	}

	// Message should always be consistent, so as to not give away which case occurred
	passwordMsg := "User or password is incorrect"

	repo := data.Repository()

	user, err := repo.GetUserByName(ctx.Request.Context(), req.Username)
	if err != nil { // Anything else
		if strings.Contains(err.Error(), "no rows") { // If user not found
			ctx.HTML(http.StatusOK, "login-message",
				gin.H{
					"type": "error",
					"message": passwordMsg})

			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 0)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword(hash, user.PasswordHash); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			ctx.HTML(http.StatusOK, "login-message",
				gin.H{
					"type": "error",
					"message": passwordMsg})

			return
		}
		l.Printf("Comparison of hash %s and %s failed: %v\n",
			hash, user.PasswordHash, err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
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
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Flash message
	if req.Password != req.Confirmation {
		ctx.HTML(http.StatusOK, "login-message",
				gin.H{
					"type": "error",
					"message": "Passwords don't match"})

		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 0)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError,
			fmt.Errorf("Password processing failed: %v", err))
		return
	} else {
		l.Printf("Registering hash %s for user %s\n", hash, req.Username)
	}

	repo := data.Repository()

	userID, err := repo.NewUser(ctx.Request.Context(),
		req.Username, []byte(req.Password))
	if err != nil {
		// Handle already existing user
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			l.Printf("Cannot register an existing user (%s)\n", req.Username)
			
			// Flash message response
			ctx.HTML(http.StatusOK, "login-message",
				gin.H{
					"type":"error",
					"message": fmt.Sprintf("User %s already exists...", req.Username)})
			
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
	setSessionCookieFromToken(ctx, newToken, 999999)
}

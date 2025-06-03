package app

import (
	"fmt"
	"lameCode/platform/data"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func LoadUserHandlers(r *gin.Engine) {
	r.GET("/login", loginPageFunc)
	r.POST("/register", registerUserFunc)
}

func loginPageFunc(ctx *gin.Context) {
	if boost := ctx.Request.Header["Hx-Boosted"]; len(boost) == 0 {
		l.Println("Printing full login template")
		l.Println(boost)
		ctx.HTML(http.StatusOK, "login.html", gin.H{})
		return
	}
	ctx.HTML(http.StatusOK, "login", gin.H{})
}

func registerUserFunc(ctx *gin.Context) {
	var req struct {
		Username     string `form:"username" binding:"required,alphanum,min=3,max=32"`
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
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	l.Printf("Created new user %s with ID %d\n", req.Username, userID)

	// Redirect to homepage
	ctx.Header("HX-Redirect", "/")
}

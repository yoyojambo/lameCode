package app

import (
	"lameCode/internal/platform/session"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Logger for the package
var l = log.New(os.Stdout, "[app] ", log.LstdFlags|log.Lmsgprefix)


type TemplateData struct {
	User UserData
	Data interface{}
}

type UserData struct {
	LoggedIn bool
	Username string
	Avatar   string
	IsAdmin  bool
}

func RenderHTML(ctx *gin.Context, code int, name string, data interface{}) {
	userData := extractUserData(ctx)

	templateData := TemplateData{
		User: userData,
		Data: data,
	}

	ctx.HTML(code, name, templateData)
}

func extractUserData(ctx *gin.Context) UserData {
	token, exists := ctx.Get(session.SessionContextKey)
	if !exists {
		return UserData{LoggedIn: false}
	}

	isAdmin, _ := ctx.Get(session.AccessContextKey)
	jwtToken := token.(*jwt.Token)

	// Extract user info from JWT claims
	claims := jwtToken.Claims.(jwt.MapClaims)
	username, _ := claims["sub"].(string)  // or however you store username
	//avatar, _ := claims["avatar"].(string) // if you store avatar in JWT

	return UserData{
		LoggedIn: true,
		Username: username,
		//Avatar:   avatar,
		IsAdmin:  isAdmin.(bool),
	}
}

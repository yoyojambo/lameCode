package app

import (
	"fmt"
	"lameCode/internal/platform/session"
	"lameCode/internal/web/ui"
	"log"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Logger for the package
var l = log.New(os.Stdout, "[app] ", log.LstdFlags|log.Lmsgprefix)


type TemplateData struct {
	User ui.UserData
	Data any
}


// RenderTempl renders a templ component with a given status code
func RenderTempl(ctx *gin.Context, status int, c templ.Component) {
	ctx.Status(status)
	if err := c.Render(ctx.Request.Context(), ctx.Writer); err != nil {
		ctx.Error(fmt.Errorf("error rendering ui template: %w", err))
	}
}

// RenderTemplOK renders a templ component with 200 OK status
func RenderTemplOK(ctx *gin.Context, c templ.Component) {
	RenderTempl(ctx, http.StatusOK, c)
}

func RenderHTML(ctx *gin.Context, code int, name string, data any) {
	userData := extractUserData(ctx)

	templateData := TemplateData{
		User: userData,
		Data: data,
	}

	ctx.HTML(code, name, templateData)
}

func extractUserData(ctx *gin.Context) ui.UserData {
	token, exists := ctx.Get(session.SessionContextKey)
	if !exists {
		return ui.UserData{LoggedIn: false}
	}

	isAdmin, _ := ctx.Get(session.AccessContextKey)
	jwtToken := token.(*jwt.Token)

	// Extract user info from JWT claims
	claims := jwtToken.Claims.(jwt.MapClaims)
	username, _ := claims["sub"].(string)  // or however you store username
	//avatar, _ := claims["avatar"].(string) // if you store avatar in JWT

	return ui.UserData{
		LoggedIn: true,
		Username: username,
		//Avatar:   avatar,
		IsAdmin:  isAdmin.(bool),
	}
}

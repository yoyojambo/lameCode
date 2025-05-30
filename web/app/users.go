package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


func LoadUserHandlers(r *gin.Engine) {
	r.GET("/login", loginPageFunc)
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

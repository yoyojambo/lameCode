package app

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoadProblemHandlers(r *gin.Engine) {
	r.GET("/problems/:id", problemFunc)
}

type User struct {
  LoggedIn  bool
  Username  string
  Avatar    string // or empty if none
}

type Problem struct {
  Title             string
  Difficulty        int
  Description   template.HTML
}

type PageData struct {
  User    User
  Problem Problem
}

func problemFunc(ctx *gin.Context) {
	data := gin.H{
		"User": User{
			LoggedIn: false,
		},

		"Problem": Problem{
			Title: "Two Sum",
			Difficulty: 1,
			Description: template.HTML("Lorem Ipsum <b>Dolor</b>"),
		},
	}
	ctx.HTML(http.StatusOK, "problem.html", data)
}

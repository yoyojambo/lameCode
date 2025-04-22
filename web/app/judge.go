package app

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Submission struct {
	Language string `form:"language" binding:"required"`
	Code     string `form:"code" binding:"required"`
}

func LoadJudgeHandlers(r *gin.Engine) {
	g := r.Group("/judge")
	g.POST("/test", printSubmission)
	g.POST("/submit", printSubmission)
}

// printSubmission is just for testing the frontend
// Response is swapped by HTMX
func printSubmission(ctx *gin.Context)  {
	var submission Submission
	if err := ctx.ShouldBind(&submission); err != nil {
		
		ctx.AbortWithStatus(http.StatusBadRequest)
		log.Println(err)
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

package main

import (
	"flag"
	"log"
	"os"

	"lameCode/platform/config"
	"lameCode/platform/data"

	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func main() {
	config.LoadServerFlags()
	flag.Parse()
	r := gin.Default()

	if config.Debug() {
		data.DB().Ping()
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	loadRoutes(r) // router.go

	port := "3000" // Default port
	// Check for override of port
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	err := r.Run(":" + port)
	log.Fatalln("Exited server with err:", err)
}

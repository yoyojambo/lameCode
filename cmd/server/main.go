package main

import (
	"flag"
	"log"
	"os"

	"lameCode/platform/config"
	"lameCode/platform/data"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadServerFlags()
	flag.Parse()

	if config.Debug() {
		data.DB().Ping()
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	loadRoutes(r) // router.go

	port := "3000" // Default port

	// Check for override of port
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	err := r.Run(":" + port)
	log.Fatalln("Exited server with err:", err)
}

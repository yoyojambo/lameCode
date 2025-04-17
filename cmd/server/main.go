package main

import (
	"flag"

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

	r.Run(":3000")
}

package main

import (
	"flag"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	
	loadRoutes(r) // router.go
	
	flag.Parse()

	r.Run(":3000")
}

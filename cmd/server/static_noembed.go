//go:build !embed_content

package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"
)

var assetsDir = flag.String("assetsDir", "./web", "Directory with location of assets/ and templates/ directories. Defaults to './web'")

// loadStaticContent loads the handlers for that static content as
// found in the folders web/assets and web/templates. In this version,
// it assumes the contents to be there.
func loadStaticContent(r *gin.Engine)  {
	log.Println("[ROUTER INIT] Loading assets from " + *assetsDir)
	r.LoadHTMLGlob(*assetsDir + "/templates/*")
	r.StaticFile("/favicon.ico", *assetsDir + "/assets/favicon.ico")
	r.Static("/assets", *assetsDir + "b/assets")
}

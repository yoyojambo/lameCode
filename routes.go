package main

import (
	"embed"
	"html/template"
	"io/fs"
	"lameCode/web/app"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed web/assets web/templates
var content embed.FS

// Loads all routes in the app, including both static assets,
// templates, and the route handlers in web/app
func loadRoutes(r *gin.Engine) {
	if gin.IsDebugging() {
		loadContentDynamically(r)
		log.Println("[ROUTER INIT] Loading assets from local files (DEV MODE)")
	} else {
		loadContentEmbedded(r)
		log.Println("[ROUTER INIT] Loading assets from files embedded at compile time")
	}

	// ALL handlers to be loaded.  
	// Should ideally only be a bunch of pkg.LoadPkgHandlers(r)
	// That does require packages to "own" their subroutes.
	//users.LoadUsersHandlers(r) // / /users/ /login
	app.LoadProblemHandlers(r)
}

// loadDynamicContent sets routes using the current working directory
// for serving static assets and templates, instead of the embedded
// content. Used for development.
func loadContentDynamically(r *gin.Engine) {
	r.LoadHTMLGlob("./web/templates/*")
	r.StaticFile("/favicon.ico", "./web/assets/favicon.ico")
	r.Static("/assets", "./web/assets")
}

// loadContentEmbedded sets routes for static assets and templates
// from an embed.FS, to more easily deploy.
// TODO: To make better use of optimizations, extract from embed and
// copy to a temp directory at startup.
func loadContentEmbedded(r *gin.Engine)  {
	// Loading templates
	templs, err := template.New("").
		ParseFS(content, "web/templates/*")
	
	if err != nil {
		log.Fatalf("[ROUTER INIT] Error parsing templates:\n%s\n", err)
	}

	r.SetHTMLTemplate(templs)

	//Loading static assets
	r.StaticFileFS("/favicon.ico", "web/assets/favicon.ico", http.FS(content))
	
	assetsFS, err := fs.Sub(content, "web/assets")
	if err != nil {
		log.Fatalf("[ROUTER INIT] Error getting subdirectory assets:\n%s\n", err)
	}

	// Loading assets from the filesystem returned by fs.Sub
	r.StaticFS("/assets", http.FS(assetsFS)) 
}


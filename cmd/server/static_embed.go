//go:build embed_content

package main

import (
	"flag"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"github.com/gin-gonic/gin"
)

//go:embed web/assets web/templates
var content embed.FS

// Optimization for embedded files, copying them
// from temp directory and reading from there.
var extractContents *bool = flag.Bool("extractStatic", false, "Extract static contents to temporary directory at runtime.")

// loadStaticContent loads the handlers for static content from an
// embed.FS, to more easily deploy.
// TODO: To make better use of optimizations, extract from embed and
// copy to a temp directory at startup.
func loadStaticContent(r *gin.Engine)  {
	l.Println("Loading assets from files embedded at compile time")
	// Loading templates
	templs, err := template.New("").
		ParseFS(content, "web/templates/*")
	
	if err != nil {
		l.Fatalf("Error parsing templates:\n%s\n", err)
	}

	r.SetHTMLTemplate(templs)

	//Loading static assets
	r.StaticFileFS("/favicon.ico", "web/assets/favicon.ico", http.FS(content))
	
	assetsFS, err := fs.Sub(content, "web/assets")
	if err != nil {
		l.Fatalf("Error getting subdirectory assets:\n%s\n", err)
	}

	// Loading assets from the filesystem returned by fs.Sub
	r.StaticFS("/assets", http.FS(assetsFS)) 
}



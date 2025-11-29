//go:build !bundle_content

package assets

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed static
var content embed.FS

// Optimization for embedded files, copying them
// from temp directory and reading from there.
var extractContents *bool = flag.Bool("extractStatic", false, "Extract static contents to temporary directory at runtime.")


var l *log.Logger = log.New(os.Stdout, "[ASSETS LOAD] ", log.LstdFlags | log.Lmsgprefix)

// loadStaticContent loads the handlers for static content from an
// embed.FS, to more easily deploy.
// TODO: To make better use of optimizations, extract from embed and
// copy to a temp directory at startup.
func LoadStaticContent(r *gin.Engine)  {
	l.Println("Loading assets from files embedded at compile time")

	//Loading static assets
	r.StaticFileFS("/favicon.ico", "static/favicon.ico", http.FS(content))
	
	assetsFS, err := fs.Sub(content, "static")
	if err != nil {
		l.Fatalf("Error getting subdirectory assets:\n%s\n", err)
	}

	// Loading assets from the filesystem returned by fs.Sub
	r.StaticFS("/assets", http.FS(assetsFS)) 
}



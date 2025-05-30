package main

import (
	"lameCode/web/app"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

var l *log.Logger = log.New(os.Stdout, "[ROUTER INIT] ", log.LstdFlags | log.Lmsgprefix)


// Loads all routes in the app, including both static assets,
// templates, and the route handlers in web/app
func loadRoutes(r *gin.Engine) {

	// Loads assets in web/assets and web/templates. Has two behaviours
	// depending on build tags:
	// 'go build .' :
	//     - Expects an 'assets' and 'templates' at runtime
	//     - Default folder is './web', changed with the --assetsDir flag
	// 'go build -tags embed_content .' :
	//     - Embeds the assets in the binary built
	//     - Will serve from the embedded content, or with flag --extractStatic
	//       makes a copy to the temporary folder and serves from that
	loadStaticContent(r)

	// ALL handlers to be loaded.  
	// Should ideally only be a bunch of pkg.LoadPkgHandlers(r)
	// That does require packages to "own" their subroutes.

	//users.LoadUsersHandlers(r) // / /users/ /login
	app.LoadProblemHandlers(r) // /problems /problem/:id
	app.LoadJudgeHandlers(r)   // /judge/test /judge/submit
	app.LoadUserHandlers(r)    // /login /register
}



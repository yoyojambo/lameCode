package main

import (
	"lameCode/internal/platform/session"
	"lameCode/internal/web/app"
	"lameCode/internal/web/assets"
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
	assets.LoadStaticContent(r)

	// ALL handlers to be loaded.  
	// Should ideally only be a bunch of pkg.LoadPkgHandlers(r)
	// That does require packages to "own" their subroutes.

	// All the dynamic handlers are expecting user-personalized
	// responses, so OptionalAuth is the default, but stricter session
	// requirements can be defined per-handler
	dynamic := r.Group("/", session.OptionalAuthRoute())
	app.LoadProblemHandlers(dynamic) // /problems /problem/:id
	app.LoadJudgeHandlers(dynamic)   // /judge/test /judge/submit
	app.LoadUserHandlers(dynamic)    // /login /register
	app.LoadUserProfileHandlers(dynamic) // /user/:username

	admin_only := r.Group("/", session.MandatoryAdminAuthRoute("/", "/login"))
	app.LoadAdminHandlers(admin_only)
	app.LoadAdminUserHandlers(admin_only)
}



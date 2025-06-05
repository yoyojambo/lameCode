package app

import (
	"lameCode/platform/config"

	"github.com/gin-gonic/gin"
)

func enableHtmxCache(ctx *gin.Context) {
	ctx.Header("Vary", "HX-Boosted")
	if !config.Debug() {
		ctx.Header("Cache-Control", "public, max-age=1800")
	} else {
		ctx.Header("Cache-Control", "max-age=600, must-revalidate")
	}
}

package session

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	SessionContextKey = "session"
	AccessContextKey  = "session-access"
	SessionCookieName = "session"
)

var (
	ErrSessionCookieNotFound = errors.New("session jwt cookie not found")
	ErrAuthRequired          = errors.New("authentication required")
	ErrAdminRequired         = errors.New("admin authorization required")
)

// Asserts that there the session cookie exists and contains a valid
// token for the current runtime and service.  The first value is the
// pointer to the parsed and verified token (if it exists), and the
// level of access (true for admin, false for others).
func assertAuthToken(ctx *gin.Context) (*jwt.Token, bool) {
	// Get cookie if it exists
	tokenStr, err := ctx.Cookie(SessionCookieName)
	if len(tokenStr) == 0 || err != nil {
		return nil, false
	}

	// Verify with current JWT secret
	token, err := VerifyJwtToken(tokenStr) // Your existing VerifyJwtToken
	if err != nil {
		// Log the underlying reason for verification failure
		ctx.Error(fmt.Errorf("JWT verification failed: %w", err))
		return nil, false // General error for middleware to check
	}

	// Check if it is an admin
	aud, err := token.Claims.GetAudience()
	if err != nil {
		ctx.Error(fmt.Errorf("Error accesing audience: %w", err))
		return token, false
	}

	// Return findings
	return token, aud[0] == "admin"
}

// Any access, but valid token
// redirect is the path to which it will redirect
func MandatoryAuthRoute(redirect string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tok, access := assertAuthToken(ctx)

		// token not found or invalid
		if tok == nil {
			// To prevent HTMX from swapping from redirect
			if ctx.GetHeader("HX-Request") == "true" {
				ctx.Header("HX-Redirect", redirect)
				ctx.Header("HX-Retarget", "body")
				ctx.AbortWithStatus(http.StatusOK)
			} else {
				ctx.Redirect(http.StatusFound, redirect)
				ctx.Abort()
			}
			return
		}

		// Set both the token and level of access in context
		ctx.Set(SessionContextKey, tok)
		ctx.Set(AccessContextKey, access)

		// Next handlers in chain
	}
}

// If there is a valid token, it is set in ctx, but it does not
// redirect anyway if there is not.
func OptionalAuthRoute() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Errors are pushed anyway by assertAuthToken
		tok, access := assertAuthToken(ctx)

		// Only set token in context if it is valid
		if tok != nil {
			ctx.Set(SessionContextKey, tok)
			ctx.Set(AccessContextKey, access)
		}
	}
}

// Allow access if authenticated as admin.
// If not authorized, redirect to user path, if not authenticated, redirect to noauth
func MandatoryAdminAuthRoute(user, noauth string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tok, access := assertAuthToken(ctx)
		if tok == nil {
			ctx.Redirect(http.StatusFound, noauth)
			ctx.Abort()
			return
		}

		ctx.Set(SessionContextKey, tok)
		ctx.Set(AccessContextKey, access)
	}
}

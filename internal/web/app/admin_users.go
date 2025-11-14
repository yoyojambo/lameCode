package app

import (
	"database/sql"
	"fmt"
	"lameCode/internal/platform/data"
	"lameCode/internal/web/ui"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
)

func LoadAdminUserHandlers(r *gin.RouterGroup) {
	// Main user management page
	r.GET("/admin/users", enableHtmxCache, adminUsersPageFunc)

	// Search endpoint
	r.GET("/admin/users/search", adminUsersSearchFunc)

	// User actions
	r.POST("/admin/users/:id/promote", adminPromoteUserFunc)
	r.POST("/admin/users/:id/demote", adminDemoteUserFunc)
	r.DELETE("/admin/users/:id", adminDeleteUserFunc)
}

func adminUsersPageFunc(ctx *gin.Context) {
	l.Println("[adminUsersPageFunc] → Rendering main user management page")

	repo := data.Repository()
	
	adminUsers, err := repo.GetUsersByStatus(ctx.Request.Context(), 1)
	if err != nil {
		err = fmt.Errorf("Error fetching Admin Users: %v", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	l.Printf("[Users Management] Loaded %d admin users", len(adminUsers))

	normalUsers, err := getUserSearchResults(ctx, "", 1)
	if err != nil {
		err = fmt.Errorf("Error fetching Normal Users Page 1: %v", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	l.Printf("[Users Management] Loaded %d normal users (Page 1)", len(normalUsers.Users))

	var adminCount, normalCount int64 = int64(len(adminUsers)), normalUsers.TotalCount

	stats := ui.UserManagementStats{
		AdminUsers:  adminCount,
		NormalUsers: normalCount,
		TotalUsers:  adminCount + normalCount,
	}

	userData := extractUserData(ctx)
	RenderTemplOK(ctx, ui.AdminUsersPage(userData, stats, adminUsers, normalUsers))
}

func adminUsersSearchFunc(ctx *gin.Context) {
	query := ctx.Query("search")
	pageStr := ctx.Query("page")

	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	result, err := getUserSearchResults(ctx, query, page)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	RenderTemplOK(ctx, ui.UserSearchResults(result))
}

func getUserSearchResults(ctx *gin.Context, query string, page int64) (ui.UserSearchResult, error) {
	const pageSize = 10
	offset := (page - 1) * pageSize

	repo := data.Repository()

	// Get total count
	totalCount, err := repo.CountUsersFiltered(ctx.Request.Context(), sql.NullString{String: query, Valid: true})
	if err != nil {
		return ui.UserSearchResult{}, err
	}

	// Get paginated users
	var users []data.User
	if query != "" {
		users, err = repo.GetUsersPaginatedFiltered(ctx.Request.Context(),
			sql.NullString{String: query, Valid: true}, pageSize+1, offset)
		if err != nil {
			return ui.UserSearchResult{}, err
		}
	} else {
		users, err = repo.GetUsersPaginated(ctx.Request.Context(), pageSize+1, offset)
	}

	// Filter out admins (only show normal users)
	normalUsers := make([]data.User, 0)
	for _, u := range users {
		if u.IsAdmin == 0 {
			normalUsers = append(normalUsers, u)
		}
	}

	// Determine pagination
	hasPrev := page > 1
	hasNext := len(normalUsers) > int(pageSize)

	// Trim to page size if we have extra
	if hasNext {
		normalUsers = normalUsers[:pageSize]
	}

	return ui.UserSearchResult{
		Users:       normalUsers,
		Query:       query,
		TotalCount:  totalCount,
		HasPrev:     hasPrev,
		HasNext:     hasNext,
		PrevPage:    page - 1,
		NextPage:    page + 1,
		CurrentPage: page,
	}, nil
}

func adminPromoteUserFunc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	repo := data.Repository()

	// Promote user to admin
	_, err = repo.UpdateUserAdminStatus(ctx.Request.Context(), 1, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Promoted user ID %d to administrator\n", id)

	// Return updated normal users table
	query := ctx.Query("search")
	page := int64(1)
	result, err := getUserSearchResults(ctx, query, page)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Render just the table body
	renderNormalUsersTable(ctx, result.Users)
}

func adminDemoteUserFunc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	repo := data.Repository()

	// Demote admin to normal user
	_, err = repo.UpdateUserAdminStatus(ctx.Request.Context(), 0, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Demoted user ID %d to normal user\n", id)

	// Return updated admin users table
	adminUsers, err := repo.GetUsersByStatus(ctx.Request.Context(), 1)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Render just the table body
	renderAdminUsersTable(ctx, adminUsers)
}

func adminDeleteUserFunc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	repo := data.Repository()

	// Check if user is admin
	user, err := repo.GetUserById(ctx.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("user not found"))
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Delete user
	err = repo.DeleteUser(ctx.Request.Context(), id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	l.Printf("Deleted user ID %d (%s)\n", id, user.Username)

	// Return updated table based on user type
	if user.IsAdmin == 1 {
		adminUsers, err := repo.GetUsersByStatus(ctx.Request.Context(), 1)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		renderAdminUsersTable(ctx, adminUsers)
	} else {
		query := ctx.Query("search")
		page := int64(1)
		result, err := getUserSearchResults(ctx, query, page)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		renderNormalUsersTable(ctx, result.Users)
	}
}

// Helper functions to render just table bodies
func renderAdminUsersTable(ctx *gin.Context, users []data.User) {
	var components []templ.Component
	for _, u := range users {
		components = append(components, ui.AdminUserRow(u))
	}

	// Render all rows
	for _, c := range components {
		c.Render(ctx.Request.Context(), ctx.Writer)
	}
}

func renderNormalUsersTable(ctx *gin.Context, users []data.User) {
	var components []templ.Component
	for _, u := range users {
		components = append(components, ui.NormalUserRow(u))
	}

	// Render all rows
	for _, c := range components {
		c.Render(ctx.Request.Context(), ctx.Writer)
	}
}

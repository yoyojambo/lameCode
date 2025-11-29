package ui

import (
	"lameCode/internal/platform/data"
)

// User data for navbar and auth state
type UserData struct {
	LoggedIn bool
	Username string
	Avatar   string
	IsAdmin  bool
}

// Stats for user management dashboard
type UserManagementStats struct {
	TotalUsers  int64
	AdminUsers  int64
	NormalUsers int64
}

// Search result with pagination
type UserSearchResult struct {
	Users       []data.User
	Query       string
	TotalCount  int64
	HasPrev     bool
	HasNext     bool
	PrevPage    int64
	NextPage    int64
	CurrentPage int64
}

// ChallengePage holds pagination data for the challenge list
type ChallengePage struct {
	Challenges  []data.Challenge
	HasPrev     bool
	HasNext     bool
	PrevPage    int64
	NextPage    int64
	CurrentPage int64
}


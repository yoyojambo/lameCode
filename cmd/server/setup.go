package main

import (
	"context"
	"crypto/rand"
	"lameCode/internal/platform/data"

	"golang.org/x/crypto/bcrypt"
)

// setupAdminUser creates an admin user in the configured database.
func setupAdminUser() {
	adminName := "admin_" + rand.Text()
	adminPass := rand.Text()
	repo := data.Repository()

	// Create new user
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), 0)
	if err != nil {
		l.Fatalf("Error creating admin account for setup: %v", err)
	}
	
	newAdmin, err := repo.NewUser(context.Background(), adminName, hash)
	if err != nil {
		l.Fatalf("Error creating admin account for setup: %v", err)
	}

	// Set as admin
	_, err = repo.UpdateUserAdminStatus(context.Background(), int64(1), newAdmin)
	
	l.Printf("\nNew admin user created!\nUSER: %s\nPASS: %s\n\n", adminName, adminPass)
}

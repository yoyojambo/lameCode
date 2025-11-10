package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"lameCode/internal/platform/data"
	"crypto/rand"
)

// GenerateBase64Password generates a password by creating a specified number of
// random bytes and encoding them into a Base64 string.
// The length parameter determines the number of random bytes used.
func GenerateBase64Password(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", fmt.Errorf("byteLength must be a positive number")
	}

	randomBytes := make([]byte, byteLength)
	rand.Read(randomBytes) // rand.Read does not return err (panics if it crashes)

	// Base64 encoding will make the string roughly (4/3) * byteLength long.
	// For example, 18 bytes -> 24 chars, 24 bytes -> 32 chars.
	encodedPassword := base64.URLEncoding.EncodeToString(randomBytes)

	return encodedPassword, nil
}

// setupAdminUser creates an admin user in the configured database.
func setupAdminUser() {
	admin_suffix := rand.Text()
	adminName := "admin_" + admin_suffix
	adminPass := rand.Text()
	repo := data.Repository()
	newAdmin, err := repo.NewUser(context.Background(), adminName, []byte(adminPass))
	if err != nil {
		l.Fatal("Error creating admin account for setup: %v", err)
	}
	_, err = repo.UpdateUserAdminStatus(context.Background(), int64(1), newAdmin)
	
	l.Printf("\nNew admin user created!\nUSER: %s\nPASS: %s\n\n", adminName, adminPass)
}

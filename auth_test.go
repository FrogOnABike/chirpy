package main

import (
	"testing"

	"github.com/frogonabike/chirpy/internal/auth"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecureP@ssw0rd!"

	// Hash the password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Error hashing password: %s", err)
	}

	// Verify the password against the hash
	match, err := auth.CheckPasswordHash(password, hashedPassword)
	if err != nil {
		t.Fatalf("Error checking password hash: %s", err)
	}
	if !match {
		t.Errorf("Password and hash do not match")
	}

	// Verify with an incorrect password
	wrongPassword := "WrongP@ssw0rd!"
	match, err = auth.CheckPasswordHash(wrongPassword, hashedPassword)
	if err != nil {
		t.Fatalf("Error checking password hash: %s", err)
	}
	if match {
		t.Errorf("Wrong password should not match the hash")
	}
}

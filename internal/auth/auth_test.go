package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecureP@ssw0rd!"

	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Error hashing password: %s", err)
	}

	// Verify the password against the hash
	match, err := CheckPasswordHash(password, hashedPassword)
	if err != nil {
		t.Fatalf("Error checking password hash: %s", err)
	}
	if !match {
		t.Errorf("Password and hash do not match")
	}

	// Verify with an incorrect password
	wrongPassword := "WrongP@ssw0rd!"
	match, err = CheckPasswordHash(wrongPassword, hashedPassword)
	if err != nil {
		t.Fatalf("Error checking password hash: %s", err)
	}
	if match {
		t.Errorf("Wrong password should not match the hash")
	}
}

func TestJWTCreationAndValidation(t *testing.T) {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	uid, _ := uuid.Parse(userID)
	tokenSecret := "testsecret"

	// Create JWT
	token, err := MakeJWT(uid, tokenSecret)
	if err != nil {
		t.Fatalf("Error creating JWT: %s", err)
	}

	// Validate JWT
	returnedUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Error validating JWT: %s", err)
	}
	if returnedUserID != uid {
		t.Errorf("Expected user ID %s, got %s", uid, returnedUserID)
	}
}

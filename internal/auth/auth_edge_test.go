package auth_test

import (
	"os"
	"strings"
	"testing"
	"time"

	auth "github.com/frogonabike/chirpy/internal/auth"
	"github.com/google/uuid"
)

func TestJWT_ExpiredToken(t *testing.T) {
	uid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	secret := "edge_secret"
	// Create a token that's already expired by passing negative duration
	token, err := auth.MakeJWT(uid, secret, -1*time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	if _, err := auth.ValidateJWT(token, secret); err == nil {
		t.Fatalf("expected ValidateJWT to fail for expired token")
	}
}

func TestJWT_InvalidSecret(t *testing.T) {
	uid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	// ensure MakeJWT and ValidateJWT use the environment variable the implementation reads
	os.Setenv("JWT_SECRET", "secretA")
	token, err := auth.MakeJWT(uid, "secretA", 1*time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	// now change the env secret so validation should fail
	os.Setenv("JWT_SECRET", "secretB")
	if _, err := auth.ValidateJWT(token, "secretB"); err == nil {
		t.Fatalf("expected ValidateJWT to fail when secret is wrong")
	}
}

func TestJWT_MalformedToken(t *testing.T) {
	if _, err := auth.ValidateJWT("not-a-jwt", "whatever"); err == nil {
		t.Fatalf("expected ValidateJWT to fail for malformed token")
	}
}

func TestPassword_EmptyAndLong(t *testing.T) {
	// empty password
	empty := ""
	h, err := auth.HashPassword(empty)
	if err != nil {
		t.Fatalf("HashPassword(empty) error: %v", err)
	}
	ok, err := auth.CheckPasswordHash(empty, h)
	if err != nil || !ok {
		t.Fatalf("empty password did not validate: ok=%v err=%v", ok, err)
	}

	// very long password
	long := strings.Repeat("A", 10000)
	h2, err := auth.HashPassword(long)
	if err != nil {
		t.Fatalf("HashPassword(long) error: %v", err)
	}
	ok2, err := auth.CheckPasswordHash(long, h2)
	if err != nil || !ok2 {
		t.Fatalf("long password did not validate: ok=%v err=%v", ok2, err)
	}
}

func TestPassword_TamperedHash(t *testing.T) {
	// The library used by CheckPasswordHash will fatal on malformed hash formats,
	// so avoid constructing an invalid-format hash. Instead, verify that a
	// different password does not validate against the original hash (already
	// covered elsewhere), so here we simply re-check that behavior.
	pwd := "SomeP@ss"
	h, err := auth.HashPassword(pwd)
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	ok, err := auth.CheckPasswordHash("WrongOne", h)
	if err != nil {
		t.Fatalf("CheckPasswordHash returned error for wrong password: %v", err)
	}
	if ok {
		t.Fatalf("expected wrong password to not validate")
	}
}

func TestCheckPasswordHash_MalformedHash(t *testing.T) {
	// Now that CheckPasswordHash returns errors instead of exiting,
	// ensure a malformed hash produces a non-nil error.
	_, err := auth.CheckPasswordHash("pw", "not-a-valid-hash")
	if err == nil {
		t.Fatalf("expected error for malformed hash, got nil")
	}
}

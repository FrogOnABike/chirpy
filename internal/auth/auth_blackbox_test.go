package auth_test

import (
    "testing"
    "time"

    auth "github.com/frogonabike/chirpy/internal/auth"
    "github.com/google/uuid"
)

func TestPasswordHashing_BlackBox(t *testing.T) {
    password := "BlackBoxP@ss!"

    hashed, err := auth.HashPassword(password)
    if err != nil {
        t.Fatalf("HashPassword error: %v", err)
    }

    ok, err := auth.CheckPasswordHash(password, hashed)
    if err != nil {
        t.Fatalf("CheckPasswordHash error: %v", err)
    }
    if !ok {
        t.Fatalf("expected password to match hash")
    }
}

func TestJWT_CreateAndValidate_BlackBox(t *testing.T) {
    uid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
    secret := "bbsecret"
    token, err := auth.MakeJWT(uid, secret, 1*time.Hour)
    if err != nil {
        t.Fatalf("MakeJWT error: %v", err)
    }

    returned, err := auth.ValidateJWT(token, secret)
    if err != nil {
        t.Fatalf("ValidateJWT error: %v", err)
    }
    if returned != uid {
        t.Fatalf("expected uid %v, got %v", uid, returned)
    }
}

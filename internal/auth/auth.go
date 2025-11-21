package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	// HashPassword generates an Argon2id hash for the provided password.
	// It returns the encoded hash string on success or a non-nil error if
	// the hashing operation fails.
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	//
	// Return values:
	//  - (true, nil)  : password matches the hash
	//  - (false, nil) : password does not match the hash
	//  - (false, err) : the hash was malformed or comparison failed
	//
	// This function returns errors instead of terminating the process so callers
	// can handle malformed hashes or other comparison errors appropriately.
	if err != nil {
		return false, err
	}
	return match, nil
}

// Function to create a JWT token for a given user ID
func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {

	// Define the signing key
	mySigningKey := []byte(tokenSecret)

	// Create the JWT claims, which includes the user ID and expiry time
	claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy",
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	// Create the token using the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	signedToken, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// Function to parse and validate a JWT token, returning the user ID if valid
func ValidateJWT(tokenString string, tokenSecret string) (uuid.UUID, error) {

	// Define the signing key
	mySigningKey := []byte(tokenSecret)

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})
	if err != nil {
		return uuid.Nil, err
	} else if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		return uuid.Parse(claims.Subject)
	} else {
		return uuid.Nil, nil
	}

}

// Function to extract Bearer token from HTTP headers
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", http.ErrNoCookie
	}

	const prefix = "Bearer "
	if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
		return "", http.ErrNoCookie
	}

	return authHeader[len(prefix):], nil

}

// Function to generate a secure random refresh token
func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("make refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// Function to extract API Key from HTTP headers
func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", http.ErrNoCookie
	}

	const prefix = "ApiKey "
	if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
		return "", http.ErrNoCookie
	}

	return authHeader[len(prefix):], nil

}

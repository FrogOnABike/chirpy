package auth

import (
	"log"
	"os"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		log.Fatal(err)
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		log.Fatal(err)
	}
	return match, nil
}

// Function to create a JWT token for a given user ID
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// Load environment variables
	godotenv.Load()

	// Get the JWT secret key from environment variables
	mySigningKey := []byte(os.Getenv("JWT_SECRET"))

	// Create the JWT claims, which includes the user ID and expiry time
	claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy",
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
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
	// Load environment variables
	godotenv.Load()

	// Get the JWT secret key from environment variables
	mySigningKey := []byte(os.Getenv("JWT_SECRET"))

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

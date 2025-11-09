package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/frogonabike/chirpy/internal/auth"
	"github.com/frogonabike/chirpy/internal/database"
)

// User creation handler - POST /api/users
func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	// Request section
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	defer r.Body.Close()

	// Validate email and password presence
	if params.Email == "" || params.Password == "" {
		respondWithError(w, 400, "Email and password are required")
		return
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		respondWithError(w, 500, "Internal server error")
		return
	}

	// Database section - prepare parameters
	dbParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	// Create user in database
	newUser, err := cfg.dbQueries.CreateUser(r.Context(), dbParams)
	if err != nil {
		respondWithError(w, 400, "Error creating user")
		return
	}

	// Map returned database user model to API user model
	createdUser := User{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}

	// Response section
	respondWithJSON(w, 201, createdUser)

}

// User login handler - POST /api/login
func (cfg *apiConfig) userLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Request section
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Expiry   int    `json:"expiry"`
	}
	expiryTime := 3600 * time.Second // Default to 1 hour

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	defer r.Body.Close()

	// Validate email and password presence
	if params.Email == "" || params.Password == "" {
		respondWithError(w, 400, "Email and password are required")
		return
	}

	// Retrieve user from database
	user, err := cfg.dbQueries.UserLogin(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 401, "incorrect email or password")
		return
	}
	// Verify password
	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, 401, "incorrect email or password")
		return
	}

	// Check and set token expiry
	if params.Expiry <= 0 || params.Expiry > 3600 {
		expiryTime = 3600 * time.Second // Default to 1 hour
	} else {
		expiryTime = time.Duration(params.Expiry) * time.Second
	}

	// Create JWT token
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expiryTime)
	if err != nil {
		log.Printf("Error creating JWT: %s", err)
		respondWithError(w, 500, "Internal server error")
		return
	}

	// Map returned database user model to API user model
	loggedInUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}

	// Response section
	respondWithJSON(w, 200, loggedInUser)
}

package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/frogonabike/chirpy/internal/auth"
	"github.com/frogonabike/chirpy/internal/database"
	"github.com/google/uuid"
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
		ChirpyRed: newUser.IsChirpyRed,
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
		// Expiry   int    `json:"expiry"`
	}
	// expiryTime := 1 * time.Hour // Default to 1 hour

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
	// if params.Expiry <= 0 || params.Expiry > 3600 {
	// 	expiryTime = 3600 * time.Second // Default to 1 hour
	// } else {
	// 	expiryTime = time.Duration(params.Expiry) * time.Second
	// }

	// Create JWT token
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret)
	if err != nil {
		log.Printf("Error creating JWT: %s", err)
		respondWithError(w, 500, "Internal server error")
		return
	}
	// Create refresh token
	refreshtoken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error creating refresh token: %s", err)
		respondWithError(w, 500, "Internal server error")
		return
	}
	// Create refresh token db record
	// dbParams for refresh token
	dbParams := database.CreateRTokenParams{
		Token:  refreshtoken,
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
	}
	_, err = cfg.dbQueries.CreateRToken(r.Context(), dbParams)
	if err != nil {
		log.Printf("Error creating refresh token: %s", err)
		respondWithError(w, 500, "Internal server error")
		return
	}

	// Map returned database user model to API user model
	loggedInUser := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshtoken,
		ChirpyRed:    user.IsChirpyRed,
	}

	// Response section
	respondWithJSON(w, 200, loggedInUser)

}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Request section
	type parameters struct {
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
	}

	// Extract JWT from Authorization header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Malformed or missing access token: %s", err)
		respondWithError(w, 401, "Missing or invalid Authorization header")
	}

	// Validate JWT and extra user ID
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		log.Printf("Invalid access token: %s", err)
		respondWithError(w, 401, "Invalid token")
	}

	// Decode request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error extracting ID from token: %s", err)
		respondWithError(w, 401, "Invalid token")
	}
	defer r.Body.Close()

	// Hash the new password
	hashedPassword, err := auth.HashPassword(params.NewPassword)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		respondWithError(w, 500, "Internal server error")
		return
	}

	// Update user in database
	updatedUser, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("Error updating user: %s", err)
		respondWithError(w, 500, "Error updating user")
		return
	}

	// Map returned database user model to API user model
	user := User{
		ID:        updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Email:     updatedUser.Email,
		ChirpyRed: updatedUser.IsChirpyRed,
	}

	// Response section
	respondWithJSON(w, 200, user)

}

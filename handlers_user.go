package main

import (
	"encoding/json"
	"log"
	"net/http"

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

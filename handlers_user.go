package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// User creation handler
func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	// Request section
	type parameters struct {
		Email string `json:"email"`
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
	// Create user in database
	newUser, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 400, "Error creating user")
		return
	}

	// Map database user to API user model
	createdUser := User{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}

	// Response section
	respondWithJSON(w, 201, createdUser)

}

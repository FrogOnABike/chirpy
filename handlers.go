package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/frogonabike/chirpy/internal/database"
	"github.com/google/uuid"
)

// Handler for readiness probe
func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

// Handler for returning server hit count
func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())))
}

// // Handler to reset metrics
// func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
// 	cfg.fileserverHits.Store(0)
// 	w.WriteHeader(200)
// }

// Handler to reset users database
func (cfg *apiConfig) resetUsersHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "User reset is only allowed in dev environment")
		return
	}
	err := cfg.dbQueries.ResetUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, "Error resetting users database")
		return
	}
	respondWithJSON(w, 200, "User database reset")
}

// Hadler to validate chirp content and create chirp
func (cfg *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	// Request section
	type parameters struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
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

	// Check if Chirp over 140 characters
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	} else {
		// If we reach here, Chirp is valid
		cleaned_body := profanityFilter(params.Body)
		params.Body = cleaned_body
	}

	// Create chirp in database
	newChirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: uuid.NullUUID{UUID: uuid.MustParse(params.UserID), Valid: true},
	})
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
	}

	// Map database chirp to API chirp model
	createdChirp := Chirp{
		ID:        newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body:      newChirp.Body,
		UserID:    newChirp.UserID.UUID,
	}

	// Response section
	respondWithJSON(w, 201, createdChirp)
}

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

package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/frogonabike/chirpy/internal/auth"
	"github.com/frogonabike/chirpy/internal/database"
	"github.com/google/uuid"
)

// Hadler to validate chirp content and create chirp - POST /api/chirps
func (cfg *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	// Request section
	type parameters struct {
		Body string `json:"body"`
	}

	// Extract JWT from Authorization header
	jwtToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Missing or invalid Authorization header")
		return
	}

	// Validate JWT
	userID, err := auth.ValidateJWT(jwtToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "Invalid token")
		return
	}

	// Decode request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
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
		UserID: uuid.NullUUID{UUID: userID, Valid: true},
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

// Handler to return all chirps
func (cfg *apiConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve all chirps from database
	chirps, err := cfg.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		log.Fatalf("Error retrieving chirps: %s", err)
		respondWithError(w, 500, "Error retrieving chirps")
		return
	}

	// Map database chirps to API chirp models
	var returnedChirps []Chirp
	for _, chirp := range chirps {
		var c Chirp
		c.ID = chirp.ID
		c.CreatedAt = chirp.CreatedAt
		c.UpdatedAt = chirp.UpdatedAt
		c.Body = chirp.Body
		c.UserID = chirp.UserID.UUID
		returnedChirps = append(returnedChirps, c)
	}
	respondWithJSON(w, 200, returnedChirps)
}

// Handler to return chirp by ID
func (cfg *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract chirpID from URL
	chirpID := r.PathValue("chirpID")
	rtnChirp, err := cfg.dbQueries.ReturnChirp(r.Context(), uuid.MustParse(chirpID))
	if err != nil {
		log.Printf("Error retrieving chirp by ID: %s", err)
		respondWithError(w, 404, "Error retrieving chirp")
		return
	}

	// Map database chirp to API chirp model
	returnedChirp := Chirp{
		ID:        rtnChirp.ID,
		CreatedAt: rtnChirp.CreatedAt,
		UpdatedAt: rtnChirp.UpdatedAt,
		Body:      rtnChirp.Body,
		UserID:    rtnChirp.UserID.UUID,
	}
	respondWithJSON(w, 200, returnedChirp)
}

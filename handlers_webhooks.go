package main

import (
	"encoding/json"
	"net/http"

	"github.com/frogonabike/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Request section

	// Check API key header
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.polkaKey {
		respondWithError(w, 401, "Missing or invalid API key")
		return
	}

	// Define expected parameters
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	// Decode request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding request body")
		return
	}
	defer r.Body.Close()

	// Process webhook event
	switch params.Event {
	case "user.upgraded":
		// Upgrade user to Chirpy Red in database
		err := cfg.dbQueries.UpgradeUserToChirpyRed(r.Context(), params.Data.UserID)
		if err != nil {
			respondWithJSON(w, 404, "")
			return
		}
		respondWithJSON(w, 204, "Processed ok")
		return
	default:
		respondWithJSON(w, 204, "No event")
		return
	}

}

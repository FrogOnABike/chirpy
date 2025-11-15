package main

import (
	"log"
	"net/http"

	"github.com/frogonabike/chirpy/internal/auth"
)

// refreshToken handler - POST /api/refresh
func (cfg *apiConfig) tokenRefreshHandler(w http.ResponseWriter, r *http.Request) {
	// Request section
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Missing or invalid Authorization header")
		return
	}

	// Validate refresh token in database
	userID, err := cfg.dbQueries.GetUserFromRToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, 401, "Invalid or expired refresh token")
		return
	}

	newJWT, err := auth.MakeJWT(userID.UUID, cfg.jwtSecret)
	if err != nil {
		log.Printf("Error creating JWT: %s", err)
		respondWithError(w, 500, "Internal server error")
		return
	}

	// Response section
	type response struct {
		Token string `json:"token"`
	}
	resp := response{
		Token: newJWT,
	}
	respondWithJSON(w, 200, resp)
}

// revokeRefreshToken handler - POST /api/revoke
func (cfg *apiConfig) revokeRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Request section
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Missing or invalid Authorization header")
		return
	}

	// Revoke refresh token in database
	err = cfg.dbQueries.RevokeRToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, 500, "Error revoking refresh token")
		return
	}
	// Response section
	respondWithJSON(w, 204, "")
}

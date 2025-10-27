package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

// Handler to reset metrics
func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(200)
}

// Hadler to validate chirp content
func vcHandler(w http.ResponseWriter, r *http.Request) {
	// Request section
	type parameters struct {
		Body string `json:"body"`
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

	// Response section

	// Check if Chirp over 140 characters
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		// respBody := returnVals{
		// 	Error: "Chirp is too long",
		// }
		// dat, err := json.Marshal(respBody)
		// if err != nil {
		// 	log.Printf("Something went wrong")
		// 	w.WriteHeader(400)
		// 	return
		// }
		// w.WriteHeader(400)
		// w.Header().Set("Content-Type", "application/json")
		// w.Write(dat)
	} else {
		// If we reach here, Chirp is valid
		cleaned_body := profanityFilter(params.Body)
		respondWithJSON(w, 200, returnVals{Cleaned_Body: cleaned_body})
		// dat, err := json.Marshal(respBody)
		// if err != nil {
		// 	log.Printf("Something went wrong")
		// 	w.WriteHeader(400)
		// 	return
		// }
		// w.WriteHeader(200)
		// w.Header().Set("Content-Type", "application/json")
		// w.Write(dat)
	}

}

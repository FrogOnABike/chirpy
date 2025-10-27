package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type returnVals struct {
	Error        string `json:"error"`
	Cleaned_Body string `json:"cleaned_body"`
}

// Helper function to respond with an error message
func respondWithError(w http.ResponseWriter, code int, msg string) {
	respBody := returnVals{
		Error: msg,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Something went wrong")
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)

}

// Helper function to respond with JSON payload
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Something went wrong")
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)
}

// Profanity filter function
func profanityFilter(body string) string {
	// List of bad words to filter, can be expanded easily
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	cleanedBody := strings.Split(body, " ")
	for i, bodyWord := range cleanedBody {
		for _, badWord := range badWords {
			if strings.ToLower(bodyWord) == badWord {
				cleanedBody[i] = "****"
			}
		}
	}
	return strings.Join(cleanedBody, " ")
}

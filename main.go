package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/frogonabike/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Connect to the database
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error connecting to database: %s", err)
	}
	defer db.Close()

	// Initialize API configuration
	apiCfg := &apiConfig{
		dbQueries: database.New(db),
	}
	apiCfg.fileserverHits.Store(0)

	// Create a new HTTP server mux
	mux := http.NewServeMux()

	// Handler to serve static files - Just to tidy up the next section :)
	fileHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))

	// Serve static files from the current directory at /app/
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileHandler))

	// Readiness probe endpoint
	mux.HandleFunc("GET /api/healthz", readyHandler)

	// Metrics endpoint
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)

	// Reset metrics endpoint
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetricsHandler)

	// Chirp validation endpoint

	mux.HandleFunc("POST /api/validate_chirp", vcHandler)
	// Start the server
	chirpyServer := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	err = chirpyServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

// Configuration struct for stateful data
type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}

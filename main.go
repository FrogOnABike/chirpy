package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/frogonabike/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Configuration struct for stateful data
type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	jwtSecret      string
	polkaKey       string
}

// *** API models - with JSON tags for serialization ***

// User model with JSON tags
type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ChirpyRed    bool      `json:"is_chirpy_red"`
}

// Chirp model with JSON tags
type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

// **** Start of the main function ****
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
		platform:  os.Getenv("PLATFORM"),
		jwtSecret: os.Getenv("JWT_SECRET"),
		polkaKey:  os.Getenv("POLKA_KEY"),
	}
	apiCfg.fileserverHits.Store(0)

	// Create a new HTTP server mux
	mux := http.NewServeMux()

	// *** General handlers ***

	// Handler to serve static files - Just to tidy up the next section :)
	fileHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))

	// Serve static files from the current directory at /app/
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileHandler))

	// Readiness probe endpoint
	mux.HandleFunc("GET /api/healthz", readyHandler)

	// Metrics endpoint
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)

	// *** User related handlers ***

	// Reset users database
	mux.HandleFunc("POST /admin/reset", apiCfg.resetUsersHandler)

	// User creation endpoint
	mux.HandleFunc("POST /api/users", apiCfg.createUserHandler)

	// User update endpoint
	mux.HandleFunc("PUT /api/users", apiCfg.updateUserHandler)

	// Login endpoint
	mux.HandleFunc("POST /api/login", apiCfg.userLoginHandler)

	// *** Chirp related handlers ***

	// Chirp creation endpoint
	mux.HandleFunc("POST /api/chirps", apiCfg.chirpHandler)

	// Return all chirps endpoint
	mux.HandleFunc("GET /api/chirps", apiCfg.getAllChirpsHandler)

	// Return specfic chirp endpoint
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpByIDHandler)

	// Delete chirp endpoint
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirpByIDHandler)

	// *** Token related handlers ***

	// Token refresh endpoint
	mux.HandleFunc("POST /api/refresh", apiCfg.tokenRefreshHandler)

	// Revoke refresh token endpoint
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeRefreshTokenHandler)

	// *** Webhook related handlers ***

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.webhookHandler)

	// *** Start the server ***
	chirpyServer := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	err = chirpyServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

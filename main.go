package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func main() {
	apiCfg := &apiConfig{}
	apiCfg.fileserverHits.Store(0)
	mux := http.NewServeMux()

	// Handler to serve static files - Just to tidy up the next section :)
	fileHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))

	// Serve static files from the current directory at /app/
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileHandler))

	// Readiness probe endpoint
	mux.HandleFunc("/healthz", readyHandler)

	// Metrics endpoint
	mux.HandleFunc("/metrics", apiCfg.metricsHandler)

	// Reset metrics endpoint
	mux.HandleFunc("/reset", apiCfg.resetMetricsHandler)

	// Start the server
	chirpyServer := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	err := chirpyServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

// Handler for readiness probe
func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

// Handler for returning server hit count
func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("Hits: %d\n", cfg.fileserverHits.Load())))
}

// Handler to reset metrics
func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(200)
}

// Configuration struct for stateful data
type apiConfig struct {
	fileserverHits atomic.Int32
}

// Middleware to increment file server hit counter
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

package main

import (
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
	mux.HandleFunc("GET /api/healthz", readyHandler)

	// Metrics endpoint
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)

	// Reset metrics endpoint
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetricsHandler)

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

// Configuration struct for stateful data
type apiConfig struct {
	fileserverHits atomic.Int32
}

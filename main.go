package main

import "net/http"

func main() {
	mux := http.NewServeMux()
	chirpyServer := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	err := chirpyServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

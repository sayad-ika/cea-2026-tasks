package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// Public
	mux.HandleFunc("GET /csrf-token", handleCSRFToken)
	mux.HandleFunc("POST /login",     handleLogin)
	mux.HandleFunc("POST /logout",    handleLogout)

	// Protected — auth + CSRF enforced
	api := http.NewServeMux()
	api.HandleFunc("GET /api/me", handleMe)
	// add more routes here: api.HandleFunc("POST /api/...", ...)

	mux.Handle("/api/", withAuth(withCSRF(api)))

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", withCORS(mux)))
}
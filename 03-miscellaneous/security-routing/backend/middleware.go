package main

import "net/http"

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getSession(r); !ok {
			jsonErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withCSRF(next http.Handler) http.Handler {
	safe := map[string]bool {"GET": true, "HEAD": true, "OPTIONS": true}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !safe[r.Method] {
			c, _ := r.Cookie("session")
			if !checkCSRF(c.Value, r.Header.Get("X-CSRF-Token")) {
				jsonErr(w, 403, "invalid CSRF token")
			    return
			}
		}
		next.ServeHTTP(w, r)
	})
}
package main

import (
	"encoding/json"
	"net/http"
)

func handleCSRFToken(w http.ResponseWriter, r *http.Request) {
	if _, ok := getSession(r); ok {
		c, _ := r.Cookie("session")
		token := randomToken()
		setCSRF(c.Value, token)
		json200(w, map[string]string{"csrfToken": token})
		return
	}

	key := randomToken()
	token := randomToken()
	setCSRF(key, token)
	setCookie(w, "preauth", key, 600)
	json200(w, map[string]string{"csrfToken": token})
}

// POST /login
func handleLogin(w http.ResponseWriter, r *http.Request) {
	preauth, err := r.Cookie("preauth")
	if err != nil || !checkCSRF(preauth.Value, r.Header.Get("X-CSRF-Token")) {
		jsonErr(w, 403, "invalid CSRF token")
		return
	}
	deleteCSRF(preauth.Value)
	setCookie(w, "preauth", "", -1)

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	usr, ok := users[body.Email]
	if !ok || usr.Password != body.Password {
		jsonErr(w, 401, "invalid credentials")
		return
	}

	sessionID := createSession(usr)
	setCookie(w, "session", sessionID, 86400)

	csrfToken := randomToken()
	setCSRF(sessionID, csrfToken)

	json200(w, map[string]any{
		"name":      usr.Name,
		"csrfToken": csrfToken,
	})
}

// POST /logout
func handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("session"); err == nil {
		deleteCSRF(c.Value)
	}
	deleteSession(w, r)
	setCookie(w, "session", "", -1)
	json200(w, map[string]string{"ok": "logged out"})
}

// GET /api/me
func handleMe(w http.ResponseWriter, r *http.Request) {
	u, _ := getSession(r)
	json200(w, map[string]any{"userId": u.ID, "name": u.Name, "email": u.Email})
}
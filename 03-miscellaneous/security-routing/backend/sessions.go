package main

import (
	"net/http"
	"sync"
)

var (
	sessionMu sync.Mutex
	sessions  = map[string]User{}
)

func createSession(user User) string {
	id := randomToken()
	sessionMu.Lock()
	sessions[id] = user
	sessionMu.Unlock()
	return id
}

func getSession(r *http.Request) (User, bool) {
	c, err := r.Cookie("session")
	if err != nil {
		return User{}, false
	}
	sessionMu.Lock()
	user, ok := sessions[c.Value]
	sessionMu.Unlock()
	return user, ok
}

func deleteSession(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("session"); err == nil {
		sessionMu.Lock()
		delete(sessions, c.Value)
		sessionMu.Unlock()
	}
}


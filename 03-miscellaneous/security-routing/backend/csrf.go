package main

import "sync"

var (
	csrfMu     sync.Mutex
	csrfTokens = map[string]string{}
)

func setCSRF(key, token string) {
	csrfMu.Lock()
	csrfTokens[key] = token
	csrfMu.Unlock()
}

func checkCSRF(key, incoming string) bool {
	csrfMu.Lock()
	stored, ok := csrfTokens[key]
	csrfMu.Unlock()
	return ok && incoming != "" && incoming == stored
}

func deleteCSRF(key string) {
	csrfMu.Lock()
	delete(csrfTokens, key)
	csrfMu.Unlock()
}
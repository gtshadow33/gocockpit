package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Session struct {
	Username  string
	CSRFToken string
	ExpiresAt time.Time
}

var (
	sessions = make(map[string]Session)
	mu       sync.RWMutex
)

func Create(username string) (string, error) {

	// Generar ID de sesión
	bytes := make([]byte, 32)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	sessionID := hex.EncodeToString(bytes)

	// Generar token CSRF
	csrfBytes := make([]byte, 32)

	_, err = rand.Read(csrfBytes)
	if err != nil {
		return "", err
	}

	csrfToken := hex.EncodeToString(csrfBytes)

	session := Session{
		Username:  username,
		CSRFToken: csrfToken,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	mu.Lock()
	sessions[sessionID] = session
	mu.Unlock()

	return sessionID, nil
}

func Get(sessionID string) (string, bool) {

	mu.RLock()
	session, exists := sessions[sessionID]
	mu.RUnlock()

	if !exists {
		return "", false
	}

	if time.Now().After(session.ExpiresAt) {

		mu.Lock()
		delete(sessions, sessionID)
		mu.Unlock()

		return "", false
	}

	return session.Username, true
}

func GetCSRFToken(sessionID string) (string, bool) {

	mu.RLock()
	session, exists := sessions[sessionID]
	mu.RUnlock()

	if !exists {
		return "", false
	}

	if time.Now().After(session.ExpiresAt) {

		mu.Lock()
		delete(sessions, sessionID)
		mu.Unlock()

		return "", false
	}

	return session.CSRFToken, true
}

func Delete(sessionID string) {

	mu.Lock()
	defer mu.Unlock()

	delete(sessions, sessionID)
}

func CleanupExpired() {

	now := time.Now()

	mu.Lock()
	defer mu.Unlock()

	for sessionID, session := range sessions {

		if now.After(session.ExpiresAt) {
			delete(sessions, sessionID)
		}
	}
}

func StartCleanup() {

	ticker := time.NewTicker(20 * time.Minute)

	go func() {

		for range ticker.C {
			CleanupExpired()
		}

	}()

}
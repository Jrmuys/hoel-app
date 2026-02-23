package server

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

type tickTickOAuthCallbackResponse struct {
	Status string `json:"status"`
}

func (s *Server) tickTickOAuthStartHandler(w http.ResponseWriter, r *http.Request) {
	if s.tickTickOAuth == nil || !s.tickTickOAuth.OAuthEnabled() {
		http.Error(w, "ticktick oauth is not configured", http.StatusBadRequest)
		return
	}

	state, err := generateOAuthState()
	if err != nil {
		http.Error(w, "unable to initialize oauth flow", http.StatusInternalServerError)
		return
	}

	s.storeTickTickOAuthState(state)

	authorizeURL, err := s.tickTickOAuth.BuildAuthorizeURL(state)
	if err != nil {
		http.Error(w, "unable to build authorize url", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authorizeURL, http.StatusFound)
}

func (s *Server) tickTickOAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if s.tickTickOAuth == nil || !s.tickTickOAuth.OAuthEnabled() {
		http.Error(w, "ticktick oauth is not configured", http.StatusBadRequest)
		return
	}

	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if state == "" || !s.consumeTickTickOAuthState(state) {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		http.Error(w, "missing oauth code", http.StatusBadRequest)
		return
	}

	if _, err := s.tickTickOAuth.ExchangeCode(r.Context(), code); err != nil {
		http.Error(w, "unable to exchange oauth code", http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, tickTickOAuthCallbackResponse{Status: "connected"})
}

func generateOAuthState() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return hex.EncodeToString(buffer), nil
}

func (s *Server) storeTickTickOAuthState(state string) {
	now := time.Now().UTC()

	s.tickTickStateLock.Lock()
	defer s.tickTickStateLock.Unlock()

	for key, expiresAt := range s.tickTickStateStore {
		if expiresAt.Before(now) {
			delete(s.tickTickStateStore, key)
		}
	}

	s.tickTickStateStore[state] = now.Add(10 * time.Minute)
}

func (s *Server) consumeTickTickOAuthState(state string) bool {
	now := time.Now().UTC()

	s.tickTickStateLock.Lock()
	defer s.tickTickStateLock.Unlock()

	expiresAt, ok := s.tickTickStateStore[state]
	if !ok {
		return false
	}

	delete(s.tickTickStateStore, state)
	return expiresAt.After(now)
}

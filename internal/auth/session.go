package auth

import (
	"sync"
	"time"
)

// Session stores a user's authentication state.
type Session struct {
	Token     string
	Lang      string
	ExpiresAt time.Time
}

// IsExpired returns true if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SessionStore defines the interface for session storage backends.
type SessionStore interface {
	Set(telegramID int64, session *Session)
	Get(telegramID int64) (*Session, bool)
	GetToken(telegramID int64) string
	GetLang(telegramID int64) string
	SetLang(telegramID int64, lang string)
	Delete(telegramID int64)
	Close() error
}

// MemoryStore provides thread-safe in-memory session storage.
type Store struct {
	mu       sync.RWMutex
	sessions map[int64]*Session
}

// NewStore creates a new session store.
func NewStore() *Store {
	return &Store{
		sessions: make(map[int64]*Session),
	}
}

// Set stores a session for a user.
func (s *Store) Set(telegramID int64, session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[telegramID] = session
}

// Get retrieves a valid session for a user.
func (s *Store) Get(telegramID int64) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[telegramID]
	if !ok || session.IsExpired() {
		return nil, false
	}
	return session, true
}

// GetToken returns the token for a user, or empty string if not found/expired.
func (s *Store) GetToken(telegramID int64) string {
	session, ok := s.Get(telegramID)
	if !ok {
		return ""
	}
	return session.Token
}

// GetLang returns the cached language for a user.
func (s *Store) GetLang(telegramID int64) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[telegramID]
	if !ok {
		return ""
	}
	return session.Lang
}

// SetLang updates the language for an existing session.
func (s *Store) SetLang(telegramID int64, lang string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, ok := s.sessions[telegramID]; ok {
		session.Lang = lang
	}
}

// Delete removes a session.
func (s *Store) Delete(telegramID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, telegramID)
}

// Close is a no-op for in-memory store (implements SessionStore interface).
func (s *Store) Close() error {
	return nil
}

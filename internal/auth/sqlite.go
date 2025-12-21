package auth

import (
	"database/sql"
	"time"
)

// SQLiteStore provides SQLite-backed session storage.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite session store.
// The db connection should already have migrations applied.
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// Set stores a session for a user.
func (s *SQLiteStore) Set(telegramID int64, session *Session) {
	_, _ = s.db.Exec(`
		INSERT OR REPLACE INTO sessions (telegram_id, token, lang, expires_at)
		VALUES (?, ?, ?, ?)
	`, telegramID, session.Token, session.Lang, session.ExpiresAt.Unix())
}

// Get retrieves a valid session for a user.
func (s *SQLiteStore) Get(telegramID int64) (*Session, bool) {
	var token, lang string
	var expiresAt int64

	err := s.db.QueryRow(`
		SELECT token, lang, expires_at FROM sessions WHERE telegram_id = ?
	`, telegramID).Scan(&token, &lang, &expiresAt)
	if err != nil {
		return nil, false
	}

	session := &Session{
		Token:     token,
		Lang:      lang,
		ExpiresAt: time.Unix(expiresAt, 0),
	}

	if session.IsExpired() {
		s.Delete(telegramID)
		return nil, false
	}

	return session, true
}

// GetToken returns the token for a user, or empty string if not found/expired.
func (s *SQLiteStore) GetToken(telegramID int64) string {
	session, ok := s.Get(telegramID)
	if !ok {
		return ""
	}
	return session.Token
}

// GetLang returns the cached language for a user.
func (s *SQLiteStore) GetLang(telegramID int64) string {
	session, ok := s.Get(telegramID)
	if !ok {
		return ""
	}
	return session.Lang
}

// SetLang updates the language for an existing session.
func (s *SQLiteStore) SetLang(telegramID int64, lang string) {
	_, _ = s.db.Exec(`UPDATE sessions SET lang = ? WHERE telegram_id = ?`, lang, telegramID)
}

// Delete removes a session.
func (s *SQLiteStore) Delete(telegramID int64) {
	_, _ = s.db.Exec(`DELETE FROM sessions WHERE telegram_id = ?`, telegramID)
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

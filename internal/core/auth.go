package core

import (
	"strconv"
	"strings"

	"github.com/archnets/telegram-bot/internal/env"
	"github.com/archnets/telegram-bot/internal/logger"
)

// AuthService handles admin authorization checks.
type AuthService struct {
	admins map[int64]struct{}
}

// NewAuthService creates a new auth service.
// The `_ any` parameter is for backwards compatibility.
func NewAuthService(_ any) *AuthService {
	s := &AuthService{
		admins: make(map[int64]struct{}),
	}

	raw := env.GetString("ADMINS_LIST", "")
	if raw == "" {
		logger.Warnf("ADMINS_LIST is empty")
		return s
	}

	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			logger.Warnf("Invalid admin ID: %q", p)
			continue
		}
		s.admins[id] = struct{}{}
	}

	logger.Infof("Loaded %d admin IDs", len(s.admins))
	return s
}

// IsAdmin checks if a telegram ID is in the admin list.
func (s *AuthService) IsAdmin(tgID int64) bool {
	_, ok := s.admins[tgID]
	return ok
}

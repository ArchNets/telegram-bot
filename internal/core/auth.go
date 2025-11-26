package core

import (
	"context"
	"fmt"

	"github.com/archnets/telegram-bot/service"
)

type AuthService struct {
	api *service.APIClient
}

func NewAuthService(api *service.APIClient) *AuthService {
	return &AuthService{api: api}
}

// LoginWithToken: handles basic validation then delegates to API client.
func (s *AuthService) LoginWithToken(ctx context.Context, tgID int64, token string) error {
	if token == "" {
		return fmt.Errorf("token is empty")
	}
	return s.api.LoginWithToken(ctx, tgID, token)
}

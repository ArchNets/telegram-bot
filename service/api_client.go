package service

import (
	"context"
	"time"
)

// APIClient will call your backend HTTP API.
// For now methods are placeholders with TODO comments.
type APIClient struct {
	BaseURL string
	Timeout time.Duration
	// You can add *http.Client here when you implement real calls.
}

func NewAPIClient(baseURL string, timeout time.Duration) *APIClient {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &APIClient{
		BaseURL: baseURL,
		Timeout: timeout,
	}
}

// LoginWithToken should call your backend to link telegram user + token.
func (c *APIClient) LoginWithToken(ctx context.Context, tgID int64, token string) error {
	// TODO: Implement a POST to something like:
	//  POST API_BASE_URL + "/telegram/login"
	//  BODY: { "telegram_id": tgID, "token": token }
	//  Handle errors based on response.
	_ = ctx
	_ = tgID
	_ = token
	return nil
}

// GetStatus should call backend and return a plain-text status for user.
func (c *APIClient) GetStatus(ctx context.Context, tgID int64) (string, error) {
	// TODO: Implement:
	//  GET API_BASE_URL + "/telegram/status/{telegram_id}"
	//  Return something like "Active until 2025-12-31" etc.
	_ = ctx
	_ = tgID
	return "Demo status from placeholder API client.\nWire this to your real backend endpoint.", nil
}

// GetConfigs should call backend and return config links / subscription info.
func (c *APIClient) GetConfigs(ctx context.Context, tgID int64) (string, error) {
	// TODO: Implement:
	//  GET API_BASE_URL + "/telegram/configs/{telegram_id}"
	//  Join links or configs into human-readable text.
	_ = ctx
	_ = tgID
	return "Demo configs from placeholder API client.\nWire this to your real configs endpoint.", nil
}

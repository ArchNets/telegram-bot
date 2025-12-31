package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles HTTP requests to the ArchNet backend API.
type Client struct {
	http    *http.Client
	baseURL string
}

// NewClient creates a new API client.
func NewClient(baseURL string, timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		http:    &http.Client{Timeout: timeout},
		baseURL: baseURL,
	}
}

// Response is the standard API response wrapper.
type Response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Error represents an API error.
type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("api error %d: %s", e.Code, e.Message)
}

// --- Request Helpers ---

// doRequest performs an HTTP request and decodes the response.
func (c *Client) doRequest(ctx context.Context, method, path string, body any, auth string) (*Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !IsSuccess(apiResp.Code) {
		return &apiResp, &Error{Code: apiResp.Code, Message: apiResp.Message}
	}

	return &apiResp, nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path, auth string) (*Response, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil, auth)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body any, auth string) (*Response, error) {
	return c.doRequest(ctx, http.MethodPost, path, body, auth)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body any, auth string) (*Response, error) {
	return c.doRequest(ctx, http.MethodPut, path, body, auth)
}

// --- User Types ---

// UserInfo represents user data from the API.
type UserInfo struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Lang      string `json:"lang"`
	ReferCode string `json:"refer_code"`
	Balance   int64  `json:"balance"`
}

// --- User API Methods ---

// GetUserInfo fetches user info using their JWT token.
func (c *Client) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
	resp, err := c.Get(ctx, EndpointUserInfo, token)
	if err != nil {
		return nil, err
	}

	var user UserInfo
	if err := json.Unmarshal(resp.Data, &user); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &user, nil
}

// UpdateUserLanguage updates the user's language preference.
func (c *Client) UpdateUserLanguage(ctx context.Context, token, lang string) error {
	_, err := c.Put(ctx, EndpointUserLang, map[string]string{"lang": lang}, token)
	return err
}

// --- Subscription Types ---

// Subscribe represents subscription plan info
type Subscribe struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Traffic     int64  `json:"traffic"`
}

// UserSubscription represents user's active subscription
type UserSubscription struct {
	ID          int64     `json:"id"`
	SubscribeID int64     `json:"subscribe_id"`
	CustomName  string    `json:"custom_name"`
	Traffic     int64     `json:"traffic"`
	Download    int64     `json:"download"`
	Upload      int64     `json:"upload"`
	ExpireTime  int64     `json:"expire_time"` // Unix milliseconds
	Status      int       `json:"status"`
	Subscribe   Subscribe `json:"subscribe"`
}

// UserSubscriptionsResponse response
type UserSubscriptionsResponse struct {
	List  []UserSubscription `json:"list"`
	Total int64              `json:"total"`
}

// GetUserSubscriptions fetches user's active subscriptions
func (c *Client) GetUserSubscriptions(ctx context.Context, token string) ([]UserSubscription, error) {
	resp, err := c.Get(ctx, EndpointUserSubscribe, token)
	if err != nil {
		return nil, err
	}

	var result UserSubscriptionsResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal subscriptions: %w", err)
	}
	return result.List, nil
}

// Login performs admin login to get backend access token.
func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
	req := map[string]string{
		"email":    email,
		"password": password,
	}
	resp, err := c.Post(ctx, "/v1/auth/login", req, "")
	if err != nil {
		return "", err
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("unmarshal login response: %w", err)
	}
	return result.Token, nil
}

// GetAuthMethodConfig fetches authentication method configuration.
func (c *Client) GetAuthMethodConfig(ctx context.Context, token, method string) (string, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/v1/admin/auth-method/config?method=%s", method), token)
	if err != nil {
		return "", err
	}

	var result struct {
		Config struct {
			BotToken string `json:"bot_token"`
		} `json:"config"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("unmarshal auth method config: %w", err)
	}
	return result.Config.BotToken, nil
}

// Package auth handles Telegram user authentication.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/archnets/telegram-bot/internal/api"
)

// TelegramUser represents user data from a Telegram update.
type TelegramUser struct {
	ID           int64
	Username     string
	FirstName    string
	LastName     string
	LanguageCode string
}

// Client handles authentication with the backend.
type Client struct {
	http     *http.Client
	baseURL  string
	botToken string
}

// NewClient creates a new auth client.
func NewClient(baseURL, botToken string) *Client {
	return &Client{
		http:     &http.Client{Timeout: 10 * time.Second},
		baseURL:  baseURL,
		botToken: botToken,
	}
}

// Authenticate authenticates a Telegram user and returns their JWT token.
func (c *Client) Authenticate(user TelegramUser) (string, error) {
	req := c.buildAuthRequest(user)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.http.Post(
		c.baseURL+api.EndpointLoginTelegram,
		"application/json",
		strings.NewReader(string(jsonData)),
	)
	if err != nil {
		return "", fmt.Errorf("auth request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth failed: status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Data.Token, nil
}

// --- Private ---

type authRequest struct {
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username,omitempty"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	Lang       string `json:"lang,omitempty"`
	Timestamp  int64  `json:"timestamp"`
	Signature  string `json:"signature"`
}

func (c *Client) buildAuthRequest(user TelegramUser) *authRequest {
	req := &authRequest{
		TelegramID: user.ID,
		Username:   user.Username,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Lang:       user.LanguageCode,
		Timestamp:  time.Now().Unix(),
	}
	req.Signature = c.computeSignature(req)
	return req
}

func (c *Client) computeSignature(req *authRequest) string {
	// Build check string with non-empty fields (sorted alphabetically)
	var fields []string

	fields = append(fields, fmt.Sprintf("auth_date=%d", req.Timestamp))
	if req.FirstName != "" {
		fields = append(fields, fmt.Sprintf("first_name=%s", req.FirstName))
	}
	fields = append(fields, fmt.Sprintf("id=%d", req.TelegramID))
	if req.LastName != "" {
		fields = append(fields, fmt.Sprintf("last_name=%s", req.LastName))
	}
	if req.Username != "" {
		fields = append(fields, fmt.Sprintf("username=%s", req.Username))
	}

	sort.Strings(fields)
	checkString := strings.Join(fields, "\n")

	// Secret = SHA256(bot_token), Signature = HMAC-SHA256(check_string, secret)
	secretKey := sha256.Sum256([]byte(c.botToken))
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(checkString))

	return hex.EncodeToString(h.Sum(nil))
}

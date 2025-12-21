// Package commands provides shared types for command handlers.
package commands

import (
	"github.com/archnets/telegram-bot/internal/api"
	"github.com/archnets/telegram-bot/internal/auth"
	"github.com/archnets/telegram-bot/internal/core"
)

// Deps contains shared dependencies for all command handlers.
type Deps struct {
	// Legacy services
	Auth         *core.AuthService
	Subscription *core.SubscriptionService

	// Config
	WebAppURL string
	BotToken  string
	BotNames  map[string]string

	// API and auth
	API             *api.Client
	AuthClient      *auth.Client
	Sessions        auth.SessionStore
	RequiredChannel string // Channel username users must join (e.g., "@Arch_Net")
}

# Coding Guidelines

Guidelines for maintaining clean, consistent code in this project.

---

## Project Structure

```
internal/
├── api/          # Backend API client, error codes, endpoints
├── auth/         # Telegram authentication & session management
├── botapp/       # Bot initialization & command routing
│   └── commands/ # Command handlers (users/, admins/)
├── core/         # Business logic (admin checks, subscriptions)
├── i18n/         # Internationalization (locales/*.json)
├── env/          # Environment variable helpers
└── logger/       # Logging utilities
```

---

## Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Files | `snake_case.go` | `handle_start.go` |
| Packages | `lowercase` | `botapp`, `i18n` |
| Public functions | `PascalCase` | `HandleStart`, `NewClient` |
| Private functions | `camelCase` | `buildAuthRequest` |
| Constants | `PascalCase` | `EndpointUserInfo` |
| Interfaces | `PascalCase` + `-er` suffix | `Authenticator` |

---

## Adding New Code

### New API Endpoint

1. Add constant to `internal/api/endpoints.go`:
   ```go
   EndpointUserBalance = "/v1/public/user/balance"
   ```

2. Add method to `internal/api/client.go`:
   ```go
   func (c *Client) GetUserBalance(ctx context.Context, token string) (int64, error) {
       resp, err := c.Get(ctx, EndpointUserBalance, token)
       // ...
   }
   ```

### New Bot Command

1. Create handler in `internal/botapp/commands/users/`:
   ```go
   // handle_balance.go
   func HandleBalance(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
       // Implementation
   }
   ```

2. Register in `internal/botapp/bot.go`:
   ```go
   register(b, "/balance", users.HandleBalance, deps)
   ```

### New Translation

1. Add key to ALL locale files (`internal/i18n/locales/*.json`):
   ```json
   "balance_info": {
       "other": "Your balance: {{.Amount}}"
   }
   ```

### New Error Code

Add to `internal/api/codes.go`:
```go
BalanceNotFound = 20011
```

---

## Middleware

Middleware wraps handlers to add reusable functionality like authentication.

### Available Middleware

| Middleware | Description |
|------------|-------------|
| `WithAuth` | Authenticates user via backend API before handler runs |
| `Chain` | Combines multiple middleware together |

### Using Middleware

Register commands with middleware in `bot.go`:

```go
// Commands that need authentication
register(b, "/status", commands.WithAuth(users.HandleStatus), deps)
register(b, "/lang", commands.WithAuth(users.HandleLanguage), deps)

// Commands with custom auth flow (no middleware)
register(b, "/start", users.HandleStart, deps)
```

### Creating New Middleware

Add to `internal/botapp/commands/middleware.go`:

```go
func WithCustom(next HandlerFunc) HandlerFunc {
    return func(ctx context.Context, b *bot.Bot, u *models.Update, deps Deps) {
        // Pre-handler logic
        next(ctx, b, u, deps)
        // Post-handler logic
    }
}
```

---

## Handler Pattern

All handlers should follow this structure:

```go
// HandleXxx handles the /xxx command.
// Note: Authentication is handled by middleware.
func HandleXxx(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
    // 1. Guard clause
    if u.Message == nil {
        return
    }

    // 2. Get language (user is already authenticated by middleware)
    lang := getLanguage(ctx, u.Message.From.ID, u.Message.From.LanguageCode, deps)
    loc := i18n.Localizer(lang)

    // 3. Business logic
    // ...

    // 4. Send response
    _, _ = b.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: u.Message.Chat.ID,
        Text:   i18n.T(loc, "message_key"),
    })
}
```

---

## Editing Code

### Do

- Extract repeated code into helper functions
- Use constants for magic strings
- Add doc comments for public functions
- Use early returns (guard clauses)
- Keep functions small (<30 lines ideally)

### Don't

- Hardcode URLs or API paths
- Duplicate translation keys
- Leave unused imports
- Mix business logic with HTTP handling

---

## Removing Code

1. **Search for usages** before deleting anything:
   ```bash
   grep -r "FunctionName" internal/
   ```

2. **Update imports** in affected files

3. **Run build** to verify:
   ```bash
   go build ./...
   ```

4. **Clean up** empty directories:
   ```bash
   find . -type d -empty -delete
   ```

---

## Code Review Checklist

- [ ] No hardcoded strings (use constants)
- [ ] Uses saved user language (not just Telegram's)
- [ ] Has error handling
- [ ] Logs important actions
- [ ] Translation keys added to ALL locales
- [ ] Builds without errors
- [ ] No unused imports

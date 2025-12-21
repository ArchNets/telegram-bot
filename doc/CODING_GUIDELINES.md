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

## Handler Pattern

All handlers should follow this structure:

```go
func HandleXxx(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
    // 1. Guard clause
    if u.Message == nil {
        return
    }

    // 2. Get logger and language
    lg := logger.ForUpdate(u)
    lang := getLanguage(ctx, u.Message.From.ID, u.Message.From.LanguageCode, deps)
    loc := i18n.Localizer(lang)

    // 3. Authenticate if needed
    if err := ensureAuthenticated(ctx, u.Message.From, deps); err != nil {
        sendError(ctx, b, u.Message.Chat.ID, lang, "auth_error")
        return
    }

    // 4. Business logic
    // ...

    // 5. Send response
    _, _ = b.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: u.Message.Chat.ID,
        Text:   i18n.T(loc, "message_key"),
    })

    // 6. Log
    lg.Infof("Command handled")
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

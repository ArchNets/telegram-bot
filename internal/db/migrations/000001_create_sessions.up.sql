CREATE TABLE IF NOT EXISTS sessions (
    telegram_id INTEGER PRIMARY KEY,
    token TEXT NOT NULL,
    lang TEXT DEFAULT '',
    expires_at INTEGER NOT NULL
);

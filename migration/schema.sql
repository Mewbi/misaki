-- Table for users
CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    telegram_id   INTEGER UNIQUE,
    telegram_name TEXT,
    admin         BOOLEAN,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

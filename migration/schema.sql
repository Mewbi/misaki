-- Table for users
CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    telegram_id   INTEGER UNIQUE,
    telegram_name TEXT,
    admin         BOOLEAN,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Table for billings
CREATE TABLE IF NOT EXISTS billings (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    value       FLOAT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- TABLE for associate billing with users
CREATE  TABLE IF NOT EXISTS billing_user (
  id_billing TEXT NOT NULL,
  id_user    TEXT NOT NULL,
  paid       BOOLEAN,
  paid_at    DATETIME,
  PRIMARY KEY (id_billing, id_user)
  FOREIGN KEY (id_billing) REFERENCES billings(id) ON DELETE CASCADE,
  FOREIGN KEY (id_user) REFERENCES users(id) ON DELETE CASCADE
);

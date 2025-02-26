CREATE SCHEMA IF NOT EXISTS main;

CREATE TABLE IF NOT EXISTS main.users (
    chat_id     BIGINT NOT NULL UNIQUE PRIMARY KEY,
    username    VARCHAR(32) NOT NULL,
    name        VARCHAR(64),
    is_admin    BOOLEAN DEFAULT FALSE,
    status      TEXT DEFAULT 'pending'
);

CREATE TABLE IF NOT EXISTS main.cameras (
    id      SERIAL PRIMARY KEY,
    name    TEXT NOT NULL UNIQUE,
    mac     TEXT NOT NULL UNIQUE
);

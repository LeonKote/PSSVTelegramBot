CREATE SCHEMA IF NOT EXISTS main;

CREATE TABLE IF NOT EXISTS main.users (
    chat_id     BIGINT NOT NULL UNIQUE PRIMARY KEY,
    username    VARCHAR(32) NOT NULL,
    name        VARCHAR(64),
    is_admin    BOOLEAN DEFAULT FALSE,
    status      TEXT DEFAULT 'pending'
);

CREATE TABLE IF NOT EXISTS main.cameras (
    name    TEXT NOT NULL UNIQUE PRIMARY KEY,
    rtsp    TEXT NOT NULL UNIQUE
);

INSERT INTO main.cameras (name, rtsp) VALUES ('camera', 'rtsp://192.168.0.100:554/ucast/11');

CREATE TABLE IF NOT EXISTS main.file_metadata (
    chat_id     BIGINT NOT NULL REFERENCES main.users(chat_id),
    camera_name TEXT NOT NULL REFERENCES main.cameras(name),
    uuid        TEXT NOT NULL UNIQUE,
    file_path   TEXT NOT NULL,
    file_size   INTEGER NOT NULL,
    file_type   TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    captured_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
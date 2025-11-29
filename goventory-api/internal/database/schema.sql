CREATE TABLE IF NOT EXISTS hosts (
    id TEXT PRIMARY KEY,
    hostname TEXT NOT NULL UNIQUE,
    ip_address TEXT NOT NULL,
    host_group TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS requests (
    id TEXT PRIMARY KEY,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    method TEXT NOT NULL,
    url TEXT NOT NULL,
    path TEXT NOT NULL,
    query_parameters TEXT,
    headers TEXT,
    body TEXT,
    response_status_code INTEGER,
    response_body TEXT,
    response_headers TEXT,
    duration_ms INTEGER,
    is_replay BOOLEAN DEFAULT 0,
    parent_id TEXT,
    replay_count INTEGER DEFAULT 0,
    tags TEXT,
    user_notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS replay_logs (
    id TEXT PRIMARY KEY,
    original_request_id TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    target_url TEXT NOT NULL,
    method TEXT NOT NULL,
    headers_sent TEXT,
    body_sent TEXT,
    response_status_code INTEGER,
    response_body TEXT,
    response_headers TEXT,
    duration_ms INTEGER,
    error TEXT,
    FOREIGN KEY (original_request_id) REFERENCES requests(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_requests_timestamp ON requests(timestamp);
CREATE INDEX IF NOT EXISTS idx_requests_method ON requests(method);
CREATE INDEX IF NOT EXISTS idx_requests_status ON requests(response_status_code);

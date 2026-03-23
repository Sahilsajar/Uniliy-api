CREATE TABLE otp_requests (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    otp_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 5,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);
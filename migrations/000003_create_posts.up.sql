CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    title TEXT,
    body TEXT,
    status TEXT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
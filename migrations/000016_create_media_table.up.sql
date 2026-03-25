CREATE TABLE media (
    id BIGSERIAL PRIMARY KEY,
    public_id TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    user_id BIGINT NOT NULL,
    is_temp BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for cleanup job
CREATE INDEX idx_media_temp_created 
ON media(is_temp, created_at);

-- Index for ownership queries
CREATE INDEX idx_media_user_id 
ON media(user_id);
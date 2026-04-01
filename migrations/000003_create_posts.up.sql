
CREATE TYPE post_status AS ENUM (
    'draft',
    'published',
    'archived',
    'deleted'
);

CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    title TEXT,
    body TEXT,
    status post_status DEFAULT 'draft',
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

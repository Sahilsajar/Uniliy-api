CREATE TABLE user_follows (
    id BIGSERIAL PRIMARY KEY,
    follower_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (follower_user_id, following_user_id),
    CHECK (follower_user_id <> following_user_id)
);
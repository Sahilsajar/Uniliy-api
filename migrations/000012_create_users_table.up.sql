CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,

    username TEXT NOT NULL CHECK (username = LOWER(TRIM(username))),
    email TEXT NOT NULL CHECK (email = LOWER(TRIM(email))),

    name TEXT,
    dob DATE,
    profile_pic TEXT,
    cover_image TEXT,
    password_hash TEXT NOT NULL,

    college_id BIGINT REFERENCES colleges(id) ON DELETE SET NULL,
    college_id_card TEXT ,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    verification_status TEXT DEFAULT 'pending',
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE (username),
    UNIQUE (email)
);

CREATE INDEX idx_users_college_id ON users (college_id);
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    name TEXT,
    dob DATE,
    profile_pic TEXT,
    cover_image TEXT,
    password_hash TEXT NOT NULL,
    college_id BIGINT REFERENCES colleges(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX users_username_unique ON users (LOWER(username));
CREATE UNIQUE INDEX users_email_unique ON users (LOWER(email));
CREATE TABLE colleges (
    id BIGSERIAL PRIMARY KEY,
    college_name TEXT NOT NULL,
    college_email TEXT,
    state TEXT,
    city TEXT,
    college_type TEXT,
    website TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
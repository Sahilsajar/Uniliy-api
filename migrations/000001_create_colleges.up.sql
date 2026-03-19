CREATE TABLE colleges (
    id BIGSERIAL PRIMARY KEY,
    college_name TEXT NOT NULL CHECK (college_name = TRIM(college_name)),
    college_email TEXT NOT NULL CHECK (college_email = LOWER(TRIM(college_email))),
    state TEXT NOT NULL,
    city TEXT NOT NULL,
    college_type TEXT,
    website_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(), 
    UNIQUE (college_email),
    UNIQUE (college_name)
);
-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auth_issuer TEXT NOT NULL,
    auth_subject TEXT NOT NULL,
    email TEXT NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    role TEXT NOT NULL CHECK (role IN ('owner')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (auth_issuer, auth_subject)
);

CREATE UNIQUE INDEX users_email_lower_unique ON users (LOWER(email));

-- +goose Down
DROP TABLE users;

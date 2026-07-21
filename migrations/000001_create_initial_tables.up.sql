CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    nickname VARCHAR(32) NOT NULL UNIQUE,
    email VARCHAR(320) NOT NULL UNIQUE CHECK (email = LOWER(email)),
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    accepted_privacy_policy BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_roles (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE registration_sessions (
    id UUID PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    nickname VARCHAR(32) NOT NULL,
    email VARCHAR(320) NOT NULL CHECK (email = LOWER(email)),
    accepted_privacy_policy BOOLEAN NOT NULL,
    code_hash CHAR(64) NOT NULL,
    code_expires TIMESTAMPTZ NOT NULL,
    attempts SMALLINT NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    last_sent_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE login_sessions (
    id UUID PRIMARY KEY,
    email VARCHAR(320) NOT NULL UNIQUE CHECK (email = LOWER(email)),
    code_hash CHAR(64) NOT NULL,
    code_expires TIMESTAMPTZ NOT NULL,
    attempts SMALLINT NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    last_sent_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE token_sessions (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash CHAR(64) NOT NULL UNIQUE,
    user_agent TEXT NOT NULL DEFAULT '',
    ip VARCHAR(45) NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_registration_sessions_email ON registration_sessions(email);
CREATE INDEX idx_registration_sessions_expires ON registration_sessions(code_expires);
CREATE INDEX idx_login_sessions_expires ON login_sessions(code_expires);
CREATE INDEX idx_token_sessions_user_id ON token_sessions(user_id);
CREATE INDEX idx_token_sessions_expires ON token_sessions(expires_at);

INSERT INTO roles (name) VALUES ('admin'), ('user');

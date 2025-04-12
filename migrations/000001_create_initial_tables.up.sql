-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    nickname VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    accepted_privacy_policy BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create user_roles table
CREATE TABLE IF NOT EXISTS user_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (user_id, role_id)
);

-- Create registration_sessions table
CREATE TABLE IF NOT EXISTS registration_sessions (
    id UUID PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    nickname VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    accepted_privacy_policy BOOLEAN NOT NULL DEFAULT false,
    code VARCHAR(6) NOT NULL,
    code_expires TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- Create login_sessions table
CREATE TABLE IF NOT EXISTS login_sessions (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    code VARCHAR(6) NOT NULL,
    code_expires TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- Create token_sessions table
CREATE TABLE IF NOT EXISTS token_sessions (
    id UUID PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    ip VARCHAR(45) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- Create indices
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_nickname ON users(nickname);
CREATE INDEX IF NOT EXISTS idx_login_sessions_email_code ON login_sessions(email, code);
CREATE INDEX IF NOT EXISTS idx_token_sessions_refresh_token ON token_sessions(refresh_token);
CREATE INDEX IF NOT EXISTS idx_token_sessions_user_id ON token_sessions(user_id);

-- Insert default roles
INSERT INTO roles (name, created_at, updated_at)
VALUES ('admin', NOW(), NOW()), ('user', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Drop indices
DROP INDEX IF EXISTS idx_token_sessions_user_id;
DROP INDEX IF EXISTS idx_token_sessions_refresh_token;
DROP INDEX IF EXISTS idx_login_sessions_email_code;
DROP INDEX IF EXISTS idx_users_nickname;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables
DROP TABLE IF EXISTS token_sessions;
DROP TABLE IF EXISTS login_sessions;
DROP TABLE IF EXISTS registration_sessions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;

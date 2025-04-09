-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TYPE outbox_status AS ENUM ('pending', 'processed', 'failed');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    hashed_password VARCHAR(255) NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE outbox (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL, 
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'::outbox_status
);

CREATE INDEX idx_users_email ON users(email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox;
DROP TABLE IF EXISTS users;

DROP INDEX IF EXISTS idx_users_email;

DROP EXTENSION IF EXISTS "uuid-ossp";
DROP TYPE IF EXISTS outbox_status;
-- +goose StatementEnd


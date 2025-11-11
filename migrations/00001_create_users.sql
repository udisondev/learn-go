-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    email VARCHAR UNIQUE NOT NULL,
    password_hash VARCHAR NOT NULL,
    registered_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    sub_plan VARCHAR NOT NULL,
    score INT NOT NULL,
    is_verified BOOLEAN NOT NULL,
    avatar_url TEXT
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_sub_plan ON users(sub_plan);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd

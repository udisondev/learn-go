-- +goose Up
-- +goose StatementBegin
ALTER TABLE email_verifications RENAME COLUMN token TO token_hash;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE email_verifications RENAME COLUMN token_hash TO token;
-- +goose StatementEnd

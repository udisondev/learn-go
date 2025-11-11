-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN phone VARCHAR;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN phone;
-- +goose StatementEnd

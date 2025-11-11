-- +goose Up
-- +goose StatementBegin
CREATE TABLE modules (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR NOT NULL,
    description TEXT NOT NULL,
    "order" INT NOT NULL,
    required_score INT NOT NULL,
    required_sub_plan VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_modules_order ON modules("order");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS modules;
-- +goose StatementEnd

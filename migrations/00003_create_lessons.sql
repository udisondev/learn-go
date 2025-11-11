-- +goose Up
-- +goose StatementBegin
CREATE TABLE lessons (
    id BIGSERIAL PRIMARY KEY,
    module_id BIGINT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    title VARCHAR NOT NULL,
    "order" INT NOT NULL,
    theory_content TEXT NOT NULL,
    required_score INT NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_lessons_module ON lessons(module_id);
CREATE INDEX idx_lessons_order ON lessons(module_id, "order");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS lessons;
-- +goose StatementEnd

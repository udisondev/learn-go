-- +goose Up
-- +goose StatementBegin
CREATE TABLE submissions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exercise_id BIGINT NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    status VARCHAR NOT NULL,
    submitted_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_submissions_user ON submissions(user_id, exercise_id);
CREATE INDEX idx_submissions_status ON submissions(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS submissions;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_progress (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exercise_id BIGINT NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    is_completed BOOLEAN NOT NULL,
    attempts INT NOT NULL,
    first_solved_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, exercise_id)
);

CREATE INDEX idx_user_progress_completed ON user_progress(user_id, is_completed);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_progress;
-- +goose StatementEnd

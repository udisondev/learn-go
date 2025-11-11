-- +goose Up
-- +goose StatementBegin
CREATE TABLE exercises (
    id BIGSERIAL PRIMARY KEY,
    lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    title VARCHAR NOT NULL,
    description TEXT NOT NULL,
    exercise_type VARCHAR NOT NULL,
    starter_code TEXT NOT NULL,
    test_cases JSONB NOT NULL,
    points INT NOT NULL,
    difficulty VARCHAR NOT NULL,
    time_limit INT NOT NULL,
    memory_limit INT NOT NULL,
    "order" INT NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_exercises_lesson ON exercises(lesson_id);
CREATE INDEX idx_exercises_difficulty ON exercises(difficulty);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS exercises;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE TABLE execution_results (
    id BIGSERIAL PRIMARY KEY,
    submission_id BIGINT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    status VARCHAR NOT NULL,
    test_results JSONB NOT NULL,
    error_message TEXT,
    execution_time INT
);

CREATE INDEX idx_execution_results_submission ON execution_results(submission_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS execution_results;
-- +goose StatementEnd

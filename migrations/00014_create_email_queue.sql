-- +goose Up
-- +goose StatementBegin
CREATE TABLE email_queue (
    id BIGSERIAL PRIMARY KEY,
    email_type TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    payload JSONB NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    status TEXT NOT NULL DEFAULT 'pending',
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    next_retry_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for efficient queue processing
-- Only index pending/processing tasks to keep index small
CREATE INDEX idx_email_queue_processing
ON email_queue(status, next_retry_at)
WHERE status IN ('pending', 'processing');

-- Index for user lookup (e.g., cancel all pending emails for a user)
CREATE INDEX idx_email_queue_user_id
ON email_queue(user_id)
WHERE user_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS email_queue;
-- +goose StatementEnd

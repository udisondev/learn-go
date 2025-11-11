-- +goose Up
-- +goose StatementBegin
CREATE TABLE achievements (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR UNIQUE NOT NULL,
    title VARCHAR NOT NULL,
    description TEXT NOT NULL,
    icon_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE user_achievements (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id BIGINT NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    earned_at TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, achievement_id)
);

CREATE INDEX idx_user_achievements_user ON user_achievements(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_achievements;
DROP TABLE IF EXISTS achievements;
-- +goose StatementEnd

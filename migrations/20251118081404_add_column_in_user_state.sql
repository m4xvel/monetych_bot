-- +goose Up
ALTER TABLE user_state DROP COLUMN review_id;

-- +goose Down

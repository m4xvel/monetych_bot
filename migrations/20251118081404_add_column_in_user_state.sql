-- +goose Up
ALTER TABLE user_state
ADD review_id INTEGER;

-- +goose Down

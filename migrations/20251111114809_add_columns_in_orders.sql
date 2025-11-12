-- +goose Up
ALTER TABLE orders
ADD COLUMN topic_id BIGINT,
ADD COLUMN thread_id INTEGER;

-- +goose Down

-- +goose Up
ALTER TABLE assessors DROP COLUMN orders_done;
ALTER TABLE user_state ADD COLUMN IF NOT EXISTS review_id BIGINT;
ALTER TABLE user_state ADD CONSTRAINT fk_user_state_reviews FOREIGN KEY (review_id) REFERENCES reviews(id) ON DELETE CASCADE;

-- +goose Down
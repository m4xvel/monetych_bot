-- +goose Up
ALTER TABLE assessors 
ADD COLUMN topic_id BIGINT UNIQUE NOT NULL;

-- +goose Down


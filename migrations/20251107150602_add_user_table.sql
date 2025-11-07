-- +goose Up
CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	tg_id BIGINT UNIQUE NOT NULL,
	is_verified BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	total_orders INT DEFAULT 0 
);
-- +goose Down


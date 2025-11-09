-- +goose Up
CREATE TABLE assessors (
	id SERIAL PRIMARY KEY,
	tg_id BIGINT UNIQUE NOT NULL,
	orders_done INT DEFAULT 0 
);
-- +goose Down

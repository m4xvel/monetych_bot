-- +goose Up
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS user_state;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS assessors;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS game_type_links;
DROP TABLE IF EXISTS game_types;
DROP TABLE IF EXISTS games;
-- +goose Down

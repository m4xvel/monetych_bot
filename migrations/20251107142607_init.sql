-- +goose Up
CREATE TABLE games (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

INSERT INTO games (name) VALUES
('Raid: Shadow Legends'),
('PUBG Mobile'),
('Genshin Impact'),
('World of Tanks'),
('DOTA 2'),
('Counter Strike 2'),
('Call of Duty: Mobile');

CREATE TABLE game_types (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

INSERT INTO game_types (name) VALUES
('Аккаунт'),
('Скин'),
('Игровая валюта'),
('Предмет'),
('Клан');

CREATE TABLE game_type_links (
    game_id INT REFERENCES games(id) ON DELETE CASCADE,
    type_id INT REFERENCES game_types(id) ON DELETE CASCADE,
    PRIMARY KEY (game_id, type_id)
);

INSERT INTO game_type_links (game_id, type_id) VALUES
(1,1),
(2,1),
(3,1),
(4,1),
(5,1),
(6,1),
(7,1);

-- +goose Down


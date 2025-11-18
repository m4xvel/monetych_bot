-- +goose Up
--------------------------------------------------
-- 1. Таблица игр
--------------------------------------------------
CREATE TABLE games (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

INSERT INTO games (name) VALUES
('Raid: Shadow Legends'),
('PUBG Mobile'),
('Genshin Impact'),
('World of Tanks'),
('DOTA 2'),
('Counter Strike 2'),
('Call of Duty: Mobile');

--------------------------------------------------
-- 2. Типы игровых товаров
--------------------------------------------------
CREATE TABLE game_types (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

INSERT INTO game_types (name) VALUES
('Аккаунт'),
('Скин'),
('Игровая валюта'),
('Предмет'),
('Клан');

--------------------------------------------------
-- 3. Связка игра → тип товара
--------------------------------------------------
CREATE TABLE game_type_links (
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    type_id BIGINT NOT NULL REFERENCES game_types(id) ON DELETE CASCADE,
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

--------------------------------------------------
-- 4. Пользователи
--------------------------------------------------
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    tg_id BIGINT UNIQUE NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    total_orders INT NOT NULL DEFAULT 0
);

--------------------------------------------------
-- 5. Оценщики
--------------------------------------------------
CREATE TABLE assessors (
    id BIGSERIAL PRIMARY KEY,
    tg_id BIGINT UNIQUE NOT NULL,
    orders_done INT NOT NULL DEFAULT 0,
    topic_id BIGINT UNIQUE
);

--------------------------------------------------
-- 6. Заказы
--------------------------------------------------
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    appraiser_id BIGINT REFERENCES assessors(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'new',
    topic_id BIGINT,
    thread_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

--------------------------------------------------
-- 7. FSM для пользователей
--------------------------------------------------
CREATE TABLE user_state (
    user_id BIGINT PRIMARY KEY,
    state VARCHAR(32) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user_state_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

--------------------------------------------------
-- 9. Отзывы
--------------------------------------------------
CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    rating SMALLINT CHECK (rating >= 1 AND rating <= 5),
    text TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_review_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    CONSTRAINT fk_review_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS reviews CASCADE;
DROP TABLE IF EXISTS user_state CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS assessors CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS game_type_links CASCADE;
DROP TABLE IF EXISTS game_types CASCADE;
DROP TABLE IF EXISTS games CASCADE;
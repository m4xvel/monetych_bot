-- +goose Up
CREATE TABLE user_state (
	user_id BIGINT PRIMARY KEY,
	state VARCHAR(32) NOT NULL,
	updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE reviews (
	id SERIAL PRIMARY KEY,
	order_id BIGINT NOT NULL,
	user_id BIGINT NOT NULL,
	rating SMALLINT CHECK (rating >= 1 AND rating <= 5),
  text TEXT,
	created_at TIMESTAMP DEFAULT now(),

	CONSTRAINT fk_review_order
	FOREIGN KEY (order_id) REFERENCES orders(id)
	ON DELETE CASCADE,

	CONSTRAINT fk_review_user
	FOREIGN KEY (user_id) REFERENCES users(id)
	ON DELETE CASCADE
);

-- +goose Down

-- +goose Up
ALTER TABLE orders
ALTER COLUMN appraiser_id DROP NOT NULL;

-- +goose Down
ALTER TABLE orders
ALTER COLUMN appraiser_id SET NOT NULL;

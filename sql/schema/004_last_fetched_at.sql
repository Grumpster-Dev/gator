-- +goose Up
ALTER TABLE feeds 
ADD COLUMN last_fetched_at TIMESTAMP WITH TIME ZONE DEFAULT NULL;


-- +goose Down
ALTER TABLE feeds 
DROP COLUMN last_fetched_at;
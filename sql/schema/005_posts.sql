-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    description TEXT,
    published_at TIMESTAMP WITH TIME ZONE,
    feed_id UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
    UNIQUE(id, feed_id)
);

-- +goose Down
DROP TABLE posts;
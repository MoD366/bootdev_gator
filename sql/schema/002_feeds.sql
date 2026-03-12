-- +goose Up

CREATE TABLE feeds (
    id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
    PRIMARY KEY (id)
);

-- +goose Down
DROP TABLE feeds;
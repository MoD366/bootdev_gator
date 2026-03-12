-- +goose Up

CREATE TABLE feed_follows (
    id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
    feed_id UUID NOT NULL REFERENCES feeds ON DELETE CASCADE,
    PRIMARY KEY (id)
);

-- +goose Down

DROP TABLE feed_follows;
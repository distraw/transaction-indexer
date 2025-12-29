-- +migrate Up
CREATE USER heathchecker;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    password BYTEA NOT NULL
);


-- +migrate Down

DROP USER heathchecker;
DROP TABLE users;
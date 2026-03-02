-- +migrate Up

CREATE USER healthchecker;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password BYTEA NOT NULL
);

CREATE TABLE addresses (
    id SERIAL PRIMARY KEY,
    addr TEXT NOT NULL UNIQUE,
    scriptpubkey TEXT NOT NULL UNIQUE
);

CREATE TABLE users_addresses (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    address_id INT REFERENCES addresses(id) ON DELETE CASCADE,
    PRIMARY KEY(user_id, address_id)
);

CREATE TABLE blocks (
    id SERIAL PRIMARY KEY,
    hash TEXT UNIQUE NOT NULL,
    previous_block_id INT REFERENCES blocks(id) ON DELETE CASCADE NULL
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    txid TEXT UNIQUE NOT NULL,
    block_id INT REFERENCES blocks(id) ON DELETE CASCADE,
    locktime BIGINT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL
);

CREATE TABLE outputs (
    id SERIAL PRIMARY KEY,

    txid TEXT NOT NULL,
    vout INT NOT NULL,
    UNIQUE(txid, vout),

    value DECIMAL NOT NULL,

    spent_in_transaction_id INT REFERENCES transactions(id) ON DELETE SET NULL,

    address TEXT NOT NULL,
    transaction_id INT REFERENCES transactions(id) ON DELETE CASCADE
);

CREATE TABLE inputs (
    id SERIAL PRIMARY KEY,

    from_address TEXT NOT NULL,
    value DECIMAL NOT NULL,

    transaction_id INT REFERENCES transactions(id) ON DELETE CASCADE,
    vin INT NOT NULL
);

-- +migrate Down

DROP USER healthchecker;
DROP TABLE users_addresses;
DROP TABLE users;
DROP TABLE transactions;
DROP TABLE inputs;
DROP TABLE outputs;
DROP TABLE addresses;
DROP TABLE blocks;
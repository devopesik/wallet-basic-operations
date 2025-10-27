-- +goose Up
CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
);

-- +goose Down
DROP TABLE IF EXISTS wallets;
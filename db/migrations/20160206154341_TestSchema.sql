
-- +goose Up
CREATE TYPE account_status AS enum('active', 'trial');
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    balance SMALLINT CHECK(balance >= 0),
    status account_status NOT NULL DEFAULT('active')
);

CREATE TABLE payments (
    id UUID PRIMARY KEY,
    account_id UUID REFERENCES accounts
);

-- +goose Down
DROP TABLE testschema;
DROP TABLE payments;
DROP TYPE account_status;

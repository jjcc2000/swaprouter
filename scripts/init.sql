CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS trades (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tx_hash     TEXT NOT NULL,
    wallet      TEXT NOT NULL,
    chain       TEXT NOT NULL,
    protocol    TEXT NOT NULL,
    from_token  TEXT NOT NULL,
    to_token    TEXT NOT NULL,
    amount_in   NUMERIC NOT NULL,
    amount_out  NUMERIC NOT NULL,
    gas_paid    NUMERIC NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trades_wallet ON trades(wallet);
CREATE INDEX idx_trades_created_at ON trades(created_at DESC);
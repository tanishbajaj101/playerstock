-- +migrate Up

CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY,
    google_sub  TEXT UNIQUE NOT NULL,
    email       TEXT NOT NULL,
    username    TEXT UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS balances (
    user_id     UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    cash        NUMERIC(20,8) NOT NULL DEFAULT 1000,
    cash_locked NUMERIC(20,8) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS assets (
    id          UUID PRIMARY KEY,
    symbol      TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS positions (
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    asset_id    UUID REFERENCES assets(id) ON DELETE CASCADE,
    qty         NUMERIC(20,8) NOT NULL DEFAULT 0,
    locked_qty  NUMERIC(20,8) NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, asset_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id),
    asset_id    UUID NOT NULL REFERENCES assets(id),
    side        SMALLINT NOT NULL,
    type        TEXT NOT NULL CHECK (type IN ('limit','market')),
    qty         NUMERIC(20,8) NOT NULL,
    filled_qty  NUMERIC(20,8) NOT NULL DEFAULT 0,
    price       NUMERIC(20,8),
    status      TEXT NOT NULL CHECK (status IN ('open','partial','filled','cancelled','rejected')),
    is_short    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS orders_user_open_idx ON orders(user_id, status) WHERE status IN ('open','partial');
CREATE INDEX IF NOT EXISTS orders_asset_idx     ON orders(asset_id, created_at DESC);

CREATE TABLE IF NOT EXISTS trades (
    id              UUID PRIMARY KEY,
    asset_id        UUID NOT NULL REFERENCES assets(id),
    buy_order_id    UUID NOT NULL REFERENCES orders(id),
    sell_order_id   UUID NOT NULL REFERENCES orders(id),
    buy_user_id     UUID NOT NULL REFERENCES users(id),
    sell_user_id    UUID NOT NULL REFERENCES users(id),
    qty             NUMERIC(20,8) NOT NULL,
    price           NUMERIC(20,8) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS trades_asset_time_idx ON trades(asset_id, created_at DESC);

CREATE TABLE IF NOT EXISTS sessions (
    id          TEXT PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL
);

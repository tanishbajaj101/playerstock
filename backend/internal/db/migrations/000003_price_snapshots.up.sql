CREATE TABLE IF NOT EXISTS price_snapshots (
    asset_id UUID          NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    ts       TIMESTAMPTZ   NOT NULL,
    price    NUMERIC(20,8) NOT NULL,
    PRIMARY KEY (asset_id, ts)
);

CREATE INDEX IF NOT EXISTS price_snapshots_asset_ts
    ON price_snapshots (asset_id, ts DESC);

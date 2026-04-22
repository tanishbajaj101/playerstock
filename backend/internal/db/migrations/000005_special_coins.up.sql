ALTER TABLE balances ADD COLUMN special_coins INTEGER NOT NULL DEFAULT 0;
UPDATE balances SET special_coins = 10;

ALTER TABLE assets ADD COLUMN supply_used INTEGER NOT NULL DEFAULT 0;

CREATE TABLE special_coin_uses (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_id   UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, asset_id)
);

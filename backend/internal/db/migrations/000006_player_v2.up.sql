ALTER TABLE assets ADD COLUMN IF NOT EXISTS date_of_birth DATE;
ALTER TABLE assets ADD COLUMN IF NOT EXISTS batting_style TEXT;
ALTER TABLE assets ADD COLUMN IF NOT EXISTS bowling_style TEXT;
ALTER TABLE assets ADD COLUMN IF NOT EXISTS player_img    TEXT;

-- CASCADE wipes positions, orders, trades, special_coin_uses, price_snapshots
TRUNCATE assets CASCADE;

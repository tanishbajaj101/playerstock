DROP TABLE IF EXISTS special_coin_uses;
ALTER TABLE assets DROP COLUMN IF EXISTS supply_used;
ALTER TABLE balances DROP COLUMN IF EXISTS special_coins;

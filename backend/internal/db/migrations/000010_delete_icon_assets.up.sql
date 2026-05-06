-- Delete assets where player_img is the generic cricapi placeholder icon.
-- Dependent rows in orders and trades have no ON DELETE CASCADE, so remove them first.
-- positions, price_snapshots, special_coin_uses cascade automatically.

DELETE FROM trades
WHERE asset_id IN (
    SELECT id FROM assets WHERE player_img = 'https://h.cricapi.com/img/icon512.png'
);

DELETE FROM orders
WHERE asset_id IN (
    SELECT id FROM assets WHERE player_img = 'https://h.cricapi.com/img/icon512.png'
);

DELETE FROM assets
WHERE player_img = 'https://h.cricapi.com/img/icon512.png';

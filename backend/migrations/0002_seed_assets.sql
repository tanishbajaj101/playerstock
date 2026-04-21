-- +migrate Up

INSERT INTO assets (id, symbol, name, description) VALUES
    ('a1000000-0000-0000-0000-000000000001', 'GOLD',   'Gold',           'Precious metal, store of value'),
    ('a1000000-0000-0000-0000-000000000002', 'SILVER',  'Silver',         'Industrial and precious metal'),
    ('a1000000-0000-0000-0000-000000000003', 'CRUDE',   'Crude Oil',      'Brent crude oil commodity'),
    ('a1000000-0000-0000-0000-000000000004', 'NIFTY',   'Nifty 50 Index', 'Indian stock market index tracker'),
    ('a1000000-0000-0000-0000-000000000005', 'DOGE2',   'Doge 2.0',       'Definitely not financial advice');

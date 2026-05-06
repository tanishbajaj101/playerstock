ALTER TABLE orders
  ADD CONSTRAINT orders_qty_integer CHECK (qty = FLOOR(qty)),
  ADD CONSTRAINT orders_qty_max     CHECK (qty <= 5),
  ADD CONSTRAINT orders_qty_min     CHECK (qty >= 1);

ALTER TABLE trades
  ADD CONSTRAINT trades_qty_integer CHECK (qty = FLOOR(qty)),
  ADD CONSTRAINT trades_qty_max     CHECK (qty <= 5),
  ADD CONSTRAINT trades_qty_min     CHECK (qty >= 1);

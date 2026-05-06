BEGIN;

CREATE TEMP TABLE players_import (
  id UUID,
  name TEXT,
  dateOfBirth TEXT,
  role TEXT,
  battingStyle TEXT,
  bowlingStyle TEXT,
  country TEXT,
  playerImg TEXT,
  team TEXT
);

\copy players_import (id, name, dateOfBirth, role, battingStyle, bowlingStyle, country, playerImg, team) FROM 'C:/stakestock/players.csv' WITH (FORMAT csv, HEADER true);

INSERT INTO assets (
  id,
  symbol,
  name,
  description,
  nationality,
  role,
  date_of_birth,
  batting_style,
  bowling_style,
  player_img,
  supply_used,
  team
)
SELECT
  p.id,
  upper(regexp_replace(trim(p.name), '[^A-Za-z0-9]+', '_', 'g')) AS symbol,
  p.name,
  '' AS description,
  NULLIF(p.country, '') AS nationality,
  NULLIF(p.role, '') AS role,
  NULLIF(p.dateOfBirth, '')::timestamp AS date_of_birth,
  NULLIF(p.battingStyle, '') AS batting_style,
  NULLIF(p.bowlingStyle, '') AS bowling_style,
  NULLIF(p.playerImg, '') AS player_img,
  0 AS supply_used,
  NULLIF(p.team, '') AS team
FROM players_import p
ON CONFLICT (id) DO UPDATE
SET
  symbol = EXCLUDED.symbol,
  name = EXCLUDED.name,
  nationality = EXCLUDED.nationality,
  role = EXCLUDED.role,
  date_of_birth = EXCLUDED.date_of_birth,
  batting_style = EXCLUDED.batting_style,
  bowling_style = EXCLUDED.bowling_style,
  player_img = EXCLUDED.player_img,
  team = EXCLUDED.team;

COMMIT;

-- Full user data reset: deletes all users and dependent data.
-- On next Google login, UpsertUser creates a fresh account → full onboarding flow.
-- Run with: psql $DATABASE_URL -f reset_users.sql

BEGIN;

DELETE FROM trades;
DELETE FROM orders;
DELETE FROM positions;
DELETE FROM sessions;
DELETE FROM balances;
DELETE FROM users;

COMMIT;

ALTER TABLE IF Exists "accounts" drop CONSTRAINT IF Exists "owner_currency_key";

ALTER TABLE IF Exists "accounts" drop CONSTRAINT IF Exists "accounts_owner_fkey";

DROP TABLE IF Exists "users";
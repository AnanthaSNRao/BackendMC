CREATE TABLE "users"(
"username" varchar PRIMARY KEY,
"hashed_password" VARCHAR NOT NULL,
"full_name" VARCHAR NOT NULL,
"email" VARCHAR NOT NULL UNIQUE,
"password_changed_at" timestamptz NOT NULL DEFAULT ('0001-01-01 00:00:00Z'),
"created_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");

ALTER TABLE "accounts" ADD CONSTRAINT "owner_currency_key" UNIQUE ("owner", "currency");

-- CREATE UNIQUE INDEX ON "accounts" ("owner", "currency");

CREATE TABLE "users" (
  "id" uuid PRIMARY KEY,
  "username" text UNIQUE NOT NULL CHECK (char_length(username) BETWEEN 3 AND 32),
  "email" text UNIQUE NOT NULL,
  "password_hash" TEXT NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT (NOW()),
  "last_login" TIMESTAMPTZ DEFAULT NULL
);
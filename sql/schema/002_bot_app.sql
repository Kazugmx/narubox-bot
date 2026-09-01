CREATE TABLE "bot_instances" (
  "id" uuid PRIMARY KEY,
  "user_id" BIGINT NOT NULL,
  "name" text NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT (NOW())
);

CREATE TABLE "channels" (
  "channel_id" text UNIQUE PRIMARY KEY CHECK (char_length(channel_id) BETWEEN 3 AND 32),
  "endpoint_id" uuid UNIQUE NOT NULL
);

CREATE TABLE "bot_subscriptions" (
  "bot_id" uuid,
  "channel_id" text NOT NULL
);



ALTER TABLE "bot_instances" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "bot_subscriptions" ADD FOREIGN KEY ("bot_id") REFERENCES "bot_instances" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "bot_subscriptions" ADD FOREIGN KEY ("channel_id") REFERENCES "channels" ("channel_id") DEFERRABLE INITIALLY IMMEDIATE;

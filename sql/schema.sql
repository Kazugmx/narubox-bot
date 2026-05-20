CREATE TABLE
    user_table (
        id INTEGER PRIMARY KEY,
        username VARCHAR(50) UNIQUE,
        mail VARCHAR(255) UNIQUE,
        password VARCHAR(100) NOT NULL,
        created_at TIMESTAMP NOT NULL,
        last_access TIMESTAMP,
        mail_token VARCHAR(100)
    );

CREATE TABLE
    bot_table (
        id uuid NOT NULL,
        owner_id INTEGER REFERENCES user_table (id),
        label VARCHAR(50) NOT NULL,
        ws_url VARCHAR(255) NOT NULL,
        mention_role_id VARCHAR(60) NOT NULL
    );

CREATE TABLE
    channel_bot_tags (
        id INTEGER PRIMARY KEY,
        channel_id VARCHAR(60) REFERENCES reg_channel (channel_id)
    );

CREATE TABLE
    reg_channel (
        channel_id VARCHAR(60) PRIMARY KEY,
        last_update TIMESTAMP,
        endpoint_id VARCHAR(90) NOT NULL
    );

    CREATE TABLE on_air(
        video_id VARCHAR(20) PRIMARY KEY,
        previous_state VARCHAR(10)
    );
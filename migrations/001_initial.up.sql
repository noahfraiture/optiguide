-- Cards
CREATE TABLE IF NOT EXISTS cards (
  id UUID PRIMARY KEY,
  idx INTEGER UNIQUE NULLS NOT DISTINCT,
  level TEXT NOT NULL,
  info TEXT,
  task_title_one TEXT,
  task_title_two TEXT,
  task_content_one TEXT,
  task_content_two TEXT,
  achievements TEXT,
  dungeon_one TEXT,
  dungeon_two TEXT,
  dungeon_three TEXT,
  spell TEXT
);

-- Users
-- The ID comes from the provider
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  provider_id TEXT NOT NULL,
  provider TEXT NOT NULL,
  username TEXT NOT NULL,
  email TEXT,
  team_size INTEGER NOT NULL,
  progress FLOAT NOT NULL,
  UNIQUE (provider_id, provider)
);

-- Users characters
-- Represent a character of the team of a user
CREATE TABLE IF NOT EXISTS user_characters (
  user_id UUID REFERENCES users (id) ON DELETE CASCADE,
  box_index INTEGER NOT NULL,
  class INTEGER NOT NULL,
  name TEXT NOT NULL,
  PRIMARY KEY (user_id, box_index)
);

-- Progress
-- Represent the progress of a user on a particular card for a particular
-- character of his team
CREATE TABLE IF NOT EXISTS progress (
  user_id UUID REFERENCES users (id) ON DELETE CASCADE,
  card_id UUID REFERENCES cards (id) ON DELETE CASCADE,
  box_index INTEGER NOT NULL,
  done BOOLEAN NOT NULL,
  PRIMARY KEY (user_id, card_id, box_index)
);

-- Guild
CREATE TABLE IF NOT EXISTS guilds (id UUID PRIMARY KEY, name VARCHAR(64) NOT NULL);

-- Guild appartenance
-- We keep it in another table than user to have the possibility of
-- having multiple guilds later
CREATE TABLE IF NOT EXISTS user_guilds (
  user_id UUID REFERENCES users (id) ON DELETE CASCADE,
  guild_id UUID REFERENCES guilds (id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, guild_id)
);

-- feedback
CREATE TABLE feedbacks (
  id SERIAL PRIMARY KEY,
  content TEXT NOT NULL,
  user_id UUID REFERENCES users (id) ON DELETE SET NULL,
  created_at TIMESTAMP
  WITH
    TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

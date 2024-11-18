-- Cards
CREATE TABLE IF NOT EXISTS cards (
  id INTEGER PRIMARY KEY,
  level INTEGER NOT NULL,
  info TEXT,
  task_merge BOOLEAN,
  task_one TEXT,
  task_two TEXT,
  achievements TEXT,
  dungeon_one TEXT,
  dungeon_two TEXT,
  dungeon_three TEXT,
  spell TEXT
);

-- Users
-- The ID comes from the provider
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  team_size INTEGER NOT NULL,
  email TEXT
);

-- Users characters
-- Represent a character of the team of a user
CREATE TABLE IF NOT EXISTS user_characters (
  user_id TEXT REFERENCES users (id) ON DELETE CASCADE,
  box_index INTEGER NOT NULL,
  class INTEGER NOT NULL,
  name TEXT NOT NULL,
  PRIMARY KEY (user_id, box_index)
);

-- Progress
-- Represent the progress of a user on a particular card for a particular
-- character of his team
CREATE TABLE IF NOT EXISTS progress (
  user_id TEXT REFERENCES users (id) ON DELETE CASCADE,
  card_id INTEGER REFERENCES cards (id) ON DELETE CASCADE,
  box_index INTEGER NOT NULL,
  done BOOLEAN NOT NULL,
  PRIMARY KEY (user_id, card_id, box_index)
);

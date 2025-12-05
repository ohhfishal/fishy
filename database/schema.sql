CREATE TABLE IF NOT EXISTS flashcards (
  header TEXT NOT NULL,
  description TEXT NOT NULL,
  origin TEXT NOT NULL,
  class_context TEXT NOT NULL,
  ai_overview TEXT,
  thumbnail TEXT,

  PRIMARY KEY (header, origin, class_context)
);


CREATE TABLE IF NOT EXISTS jobs (
  id INTEGER PRIMARY KEY,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,

  -- 0: Success, 1: This failed, 2: This and previous failed....
  failures INTEGER NOT NULL
);

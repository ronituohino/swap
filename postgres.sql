CREATE DATABASE search_index;

CREATE TABLE websites (
  id SERIAL PRIMARY KEY, -- SERIAL = 4 bytes, 1 to 2,147,483,647 range, plenty for this project
  url TEXT NOT NULL, -- max length is 2048 chars, use text because simple
  title TEXT NOT NULL, -- max length is 512 chars
  description TEXT NOT NULL
);

CREATE TABLE keywords (
  id SERIAL PRIMARY KEY,
  word TEXT UNIQUE NOT NULL
);

CREATE TABLE relations (
  id SERIAL PRIMARY KEY,
  website_id INT NOT NULL REFERENCES websites(id),
  keyword_id INT NOT NULL REFERENCES keywords(id),
  relevance INT NOT NULL -- our own metric
);

-- TODO: CREATE INDEX ...
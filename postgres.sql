CREATE DATABASE search_index;

CREATE TABLE websites (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  url VARCHAR(2048) NOT NULL, -- max length is 2048 chars
  title VARCHAR(512) NOT NULL, -- max length is 512 chars
  description VARCHAR(512) NOT NULL -- no official max length, but limit to 512 chars
);

CREATE TABLE keywords (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  word VARCHAR(64) UNIQUE NOT NULL
);

CREATE TABLE relations (
  id BIGSERIAL PRIMARY KEY,
  website_id UUID NOT NULL REFERENCES websites(id),
  keyword_id UUID NOT NULL REFERENCES keywords(id),
  relevance INT NOT NULL -- our own metric
);

-- TODO: CREATE INDEX ...
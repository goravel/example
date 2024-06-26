CREATE TABLE users (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) DEFAULT '' NOT NULL,
  avatar varchar(255) DEFAULT '' NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);

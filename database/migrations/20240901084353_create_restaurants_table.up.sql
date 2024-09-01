CREATE TABLE restaurants (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) DEFAULT '' NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);

CREATE TABLE restaurant_categories (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  restaurantId integer NOT NULL,
  categoryId integer NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);

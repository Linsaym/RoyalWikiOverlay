CREATE TABLE IF NOT EXISTS monsters (
                                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                                        name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS items (
                                     id INTEGER PRIMARY KEY AUTOINCREMENT,
                                     name TEXT NOT NULL,
                                     type TEXT NOT NULL,
                                     price INTEGER NOT NULL,
                                     wiki_url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS locations (
                                         id INTEGER PRIMARY KEY AUTOINCREMENT,
                                         name TEXT NOT NULL UNIQUE,
                                         min_level INTEGER NOT NULL,
                                         max_level INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS monster_items (
                                             monster_id INTEGER NOT NULL,
                                             item_id INTEGER NOT NULL,
                                             PRIMARY KEY (monster_id, item_id)
    );

CREATE TABLE IF NOT EXISTS monster_locations (
                                                 monster_id INTEGER NOT NULL,
                                                 location_id INTEGER NOT NULL,
                                                 PRIMARY KEY (monster_id, location_id)
    );

CREATE UNIQUE INDEX IF NOT EXISTS idx_items_name_unique
    ON items (name);
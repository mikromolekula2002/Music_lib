-- Создание таблицы songs_info
CREATE TABLE IF NOT EXISTS song_info (
    id SERIAL PRIMARY KEY,
    group_name VARCHAR(255) NOT NULL,
    song VARCHAR(255) NOT NULL,
    release_date DATE NOT NULL,
    link VARCHAR(500) NOT NULL,
    UNIQUE (group_name, song) -- Запрещаем дубликаты одной песни у одной группы
);

-- Создание таблицы song_lyrics
CREATE TABLE IF NOT EXISTS song_text (
    id SERIAL PRIMARY KEY,
    song_id INT NOT NULL REFERENCES song_info(id) ON DELETE CASCADE,
    verse TEXT NOT NULL,
);

-- Индекс для быстрого поиска song_id по group_name + song
CREATE INDEX idx_songs_group_song ON song_info (group_name, song);

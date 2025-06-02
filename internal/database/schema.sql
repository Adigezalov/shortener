-- Создаем таблицу для хранения URL
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_id VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создаем индекс для быстрого поиска по original_url
CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls (original_url);

-- Создаем индекс для быстрого поиска по short_id
CREATE INDEX IF NOT EXISTS idx_urls_short_id ON urls (short_id); 
-- Создаем таблицу для хранения URL
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_id VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    user_id VARCHAR(36),
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создаем уникальный индекс для original_url
CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_original_url_unique ON urls (original_url);

-- Создаем индекс для быстрого поиска по short_id
CREATE INDEX IF NOT EXISTS idx_urls_short_id ON urls (short_id);

-- Создаем индекс для поиска URL пользователя
CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls (user_id); 
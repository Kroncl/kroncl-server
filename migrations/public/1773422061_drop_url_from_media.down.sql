-- Down Migration: drop_url_from_media
-- Type: public
-- Created: 2026-03-13 20:15:00

-- Возвращаем колонку url (для отката)
ALTER TABLE media_files ADD COLUMN url text;

-- Делаем колонку NOT NULL
ALTER TABLE media_files ALTER COLUMN url SET NOT NULL;

-- Возвращаем комментарии
COMMENT ON TABLE media_files IS 'Глобальное хранилище файлов (аватары и прочий хлам)';
COMMENT ON COLUMN media_files.path IS 'Путь в MinIO';
COMMENT ON COLUMN media_files.url IS 'Публичный URL для доступа';
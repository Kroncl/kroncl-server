-- Up Migration: drop_url_from_media
-- Type: public
-- Created: 2026-03-13 20:15:00

-- Удаляем колонку url, так как URL теперь генерируются на лету через presigned URLs
ALTER TABLE media_files DROP COLUMN IF EXISTS url;

-- Обновляем комментарии
COMMENT ON TABLE media_files IS 'Глобальное хранилище файлов (аватары и прочий хлам). URL генерируется на лету через presigned URLs.';
COMMENT ON COLUMN media_files.path IS 'Путь в MinIO (используется для генерации presigned URL)';
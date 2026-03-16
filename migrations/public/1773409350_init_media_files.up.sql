-- Up Migration: init_media_files
-- Type: public
-- Created: 2026-03-13 16:42:30

CREATE TABLE media_files (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    path text NOT NULL,                    -- путь в MinIO (avatars/uuid.jpg)
    url text NOT NULL,                      -- полный URL для доступа
    size bigint NOT NULL,                    -- размер в байтах
    mime_type varchar(100) NOT NULL,         -- image/jpeg, image/png
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    
    -- Метаданные
    original_name varchar(255),               -- оригинальное имя файла
    metadata jsonb DEFAULT '{}'::jsonb,
    
    -- Для связей с другими сущностями (опционально)
    entity_type varchar(50),                   -- 'account_avatar', 'company_logo', etc
    entity_id uuid                              -- ID сущности
);

-- Индексы
CREATE INDEX idx_media_files_created_by ON media_files(created_by);
CREATE INDEX idx_media_files_entity ON media_files(entity_type, entity_id);
CREATE INDEX idx_media_files_created_at ON media_files(created_at DESC);

-- Комментарии
COMMENT ON TABLE media_files IS 'Глобальное хранилище файлов (аватары и прочий хлам)';
COMMENT ON COLUMN media_files.path IS 'Путь в MinIO';
COMMENT ON COLUMN media_files.url IS 'Публичный URL для доступа';
COMMENT ON COLUMN media_files.created_by IS 'ID аккаунта, загрузившего файл';
COMMENT ON COLUMN media_files.entity_type IS 'Тип сущности, к которой привязан файл';
COMMENT ON COLUMN media_files.entity_id IS 'ID сущности';
-- Up Migration: add_unique_index_for_email
-- Created: 2025-12-04 18:36:22

-- Сначала чистим возможные дубликаты (если они есть)
-- Удаляем все кроме самой первой записи для каждого email
DELETE FROM accounts a1
USING accounts a2
WHERE a1.email = a2.email
  AND LOWER(a1.email) = LOWER(a2.email)
  AND a1.id > a2.id;

-- Добавляем уникальный индекс на email (регистронезависимый)
CREATE UNIQUE INDEX unique_lower_email 
ON accounts (LOWER(email));

-- Или, если хотите обычный уникальный constraint:
-- ALTER TABLE accounts ADD CONSTRAINT unique_email UNIQUE (email);

-- Комментарий к ограничению
COMMENT ON INDEX unique_lower_email IS 'Гарантирует уникальность email независимо от регистра';

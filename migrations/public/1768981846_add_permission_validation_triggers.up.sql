-- Up Migration: add_permission_cascade_delete_triggers
-- Created: 2026-01-21 11:00:00

-- 1. Создаем функцию для каскадного удаления permissions
CREATE OR REPLACE FUNCTION cascade_delete_permission()
RETURNS TRIGGER AS $$
BEGIN
    -- Удаляем permission из всех ролей (JSONB array)
    UPDATE roles 
    SET permissions = permissions - OLD.code::text
    WHERE permissions @> jsonb_build_array(OLD.code);
    
    -- Удаляем permission из всех company_accounts (JSONB object)
    UPDATE company_accounts 
    SET permissions = permissions - OLD.code::text
    WHERE permissions ? OLD.code::text;
    
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- 2. Создаем триггер для каскадного удаления
DROP TRIGGER IF EXISTS cascade_delete_permission ON permissions;
CREATE TRIGGER cascade_delete_permission
    BEFORE DELETE ON permissions
    FOR EACH ROW
    EXECUTE FUNCTION cascade_delete_permission();

-- 3. Функция валидации (оставляем как есть)
CREATE OR REPLACE FUNCTION validate_permissions()
RETURNS TRIGGER AS $$
DECLARE
    perm_code TEXT;
    perm_exists BOOLEAN;
BEGIN
    IF NEW.permissions IS NOT NULL THEN
        -- Проверяем JSONB массив (для roles.permissions)
        IF jsonb_typeof(NEW.permissions) = 'array' THEN
            FOR perm_code IN 
                SELECT value::TEXT 
                FROM jsonb_array_elements_text(NEW.permissions)
            LOOP
                SELECT EXISTS(
                    SELECT 1 FROM permissions WHERE code = perm_code
                ) INTO perm_exists;
                
                IF NOT perm_exists THEN
                    RAISE EXCEPTION 'Permission code "%" does not exist in permissions table', perm_code;
                END IF;
            END LOOP;
        
        -- Проверяем JSONB объект (для company_accounts.permissions)
        ELSIF jsonb_typeof(NEW.permissions) = 'object' THEN
            FOR perm_code IN 
                SELECT key 
                FROM jsonb_each_text(NEW.permissions)
            LOOP
                SELECT EXISTS(
                    SELECT 1 FROM permissions WHERE code = perm_code
                ) INTO perm_exists;
                
                IF NOT perm_exists THEN
                    RAISE EXCEPTION 'Permission code "%" does not exist in permissions table', perm_code;
                END IF;
            END LOOP;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 4. Создаем триггеры для валидации
DROP TRIGGER IF EXISTS validate_role_permissions ON roles;
CREATE TRIGGER validate_role_permissions
    BEFORE INSERT OR UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION validate_permissions();

DROP TRIGGER IF EXISTS validate_account_permissions ON company_accounts;
CREATE TRIGGER validate_account_permissions
    BEFORE INSERT OR UPDATE ON company_accounts
    FOR EACH ROW
    EXECUTE FUNCTION validate_permissions();
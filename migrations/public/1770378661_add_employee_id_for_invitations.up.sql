-- Up Migration: add_employee_id_for_invitations
-- Type: public
-- Created: 2026-02-06 14:51:01

-- Добавляем поле employee_id для связи с сотрудниками компании
ALTER TABLE company_invitations 
ADD COLUMN employee_id UUID NULL;

-- Добавляем комментарий к новому полю
COMMENT ON COLUMN company_invitations.employee_id IS 'ID сотрудника в компании (связь будет настроена позже)';

-- Создаем индекс для оптимизации запросов по employee_id
CREATE INDEX idx_company_invitations_employee_id 
ON company_invitations(employee_id) 
WHERE employee_id IS NOT NULL;
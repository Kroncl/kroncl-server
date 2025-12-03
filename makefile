.PHONY: migrate-up migrate-down migrate-create migrate-version

MIGRATE_CMD = go run cmd/migrate/main.go

migrate-up:
	$(MIGRATE_CMD) -cmd up

migrate-down:
	$(MIGRATE_CMD) -cmd down

migrate-version:
	$(MIGRATE_CMD) -cmd version

migrate-create:
	@read -p "Введите имя миграции: " name; \
	$(MIGRATE_CMD) -cmd create -name $$name

migrate-steps:
	@read -p "Количество шагов (+ вверх, - вниз): " steps; \
	$(MIGRATE_CMD) -cmd steps -steps $$steps
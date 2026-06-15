# Базовые настройки
DC = docker-compose

# Docker команды
ps:
	@$(DC) ps

up:
	@$(DC) up -d

down:
	@$(DC) down

stop:
	@$(DC) stop

restart:
	@$(DC) down
	@$(DC) up -d

build:
	@$(DC) down
	@$(DC) up --build -d

# Справка
help:
	@echo "Docker команды:"
	@echo "  ps        - Показать контейнеры"
	@echo "  up        - Запустить контейнеры"
	@echo "  down      - Остановить и удалить"
	@echo "  stop      - Остановить контейнеры"
	@echo "  restart   - Перезапустить"
	@echo "  build     - Пересобрать и запустить"
	@echo ""
	@echo "Тестовые команды:"
	@echo "  test-setup    - Настройка тестовой БД"
	@echo "  test          - Запустить тесты"
	@echo "  test-fresh    - Сбросить БД и тесты"
	@echo "  migrate-test  - Миграции для тестов"
	@echo "  db-check      - Проверить БД"

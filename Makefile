# Установка зависимостей
deps:
	go mod tidy
	go mod download

# Миграции
migrate-up:
	@echo "Применение миграций..."
	@if [ -n "$(DATABASE_DSN)" ]; then \
		export PATH=$$PATH:$$(go env GOPATH)/bin && migrate -path migrations -database "$(DATABASE_DSN)" up; \
	else \
		echo "Ошибка: DATABASE_DSN не установлен"; \
		exit 1; \
	fi

migrate-down:
	@echo "Откат миграций..."
	@if [ -n "$(DATABASE_DSN)" ]; then \
		export PATH=$$PATH:$$(go env GOPATH)/bin && migrate -path migrations -database "$(DATABASE_DSN)" down; \
	else \
		echo "Ошибка: DATABASE_DSN не установлен"; \
		exit 1; \
	fi

migrate-force:
	@echo "Принудительная установка версии миграции..."
	@if [ -n "$(DATABASE_DSN)" ] && [ -n "$(VERSION)" ]; then \
		export PATH=$$PATH:$$(go env GOPATH)/bin && migrate -path migrations -database "$(DATABASE_DSN)" force $(VERSION); \
	else \
		echo "Ошибка: DATABASE_DSN или VERSION не установлены"; \
		exit 1; \
	fi

migrate-version:
	@echo "Текущая версия миграции..."
	@if [ -n "$(DATABASE_DSN)" ]; then \
		export PATH=$$PATH:$$(go env GOPATH)/bin && migrate -path migrations -database "$(DATABASE_DSN)" version; \
	else \
		echo "Ошибка: DATABASE_DSN не установлен"; \
		exit 1; \
	fi


migrate-create:
	@echo "Создание новой миграции..."
	@if [ -n "$(NAME)" ]; then \
  		echo "Создание миграции для Postgres..."; \
		export PATH=$$PATH:$$(go env GOPATH)/bin && migrate create -ext sql -dir migrations -seq $(NAME); \
		echo "Миграции созданы успешно!"; \
	else \
		echo "Ошибка: NAME не установлен. Используйте: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi

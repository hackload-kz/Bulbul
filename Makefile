.PHONY: run test build clean deploy deploy-api deploy-infra build-linux test-build help all

# Default target for deployment
all: build-linux deploy

# Запуск сервера
run:
	go run ./cmd/api

# Запуск тестов
test:
	go test -v ./...

# Валидация API на соответствие спецификации
validate:
	go run scripts/validate.go

# Валидация с кастомным URL
validate-url:
	@read -p "Enter API URL: " url; \
	go run scripts/validate.go -url $$url

# Сборка приложения (локально)
build:
	go build -o bulbul ./cmd/api

# Build the API server binary locally for Linux deployment
build-linux:
	@echo "Building API server binary for Linux deployment..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
          -ldflags="-w -s" \
          -o "./infra/files/api-server" \
          ./cmd/api

# Очистка
clean:
	rm -f bulbul infra/files/api-server

# Установка зависимостей
deps:
	go mod tidy
	go mod download

# Проверка кода
lint:
	golangci-lint run

# Форматирование кода
fmt:
	go fmt ./...
	go vet ./...

.PHONY: run test build clean

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

# Сборка приложения
build:
	go build -o bulbul .

# Очистка
clean:
	rm -f bulbul

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

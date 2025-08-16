# Bulbul
HackLoad 2025 - Репозиторий команды Булбул


# How to run
Для начала работы нужен "docker compose up -d"

После можно загрузить данные: ./load-data.sh (это зальет users.sql and events.sql в базу, сам скачает)

Запуск api: go run cmd/api/main.go
Запуск  consumers: go run cmd/consumers/main.go
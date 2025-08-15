package main

import (
	"log"
	"os"

	"bulbul/internal/api"
	"bulbul/internal/config"
	"bulbul/internal/validation"
)

func main() {
	// Проверяем, нужно ли запустить валидацию
	if len(os.Args) > 1 && os.Args[1] == "validate" {
		validation.RunValidation()
		return
	}

	// Загружаем конфигурацию
	cfg := config.Load()

	// Создаем и настраиваем сервер
	server := api.NewServer(cfg)

	// Запускаем сервер
	log.Printf("Starting server on :%s", cfg.Port)
	if err := server.Run(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

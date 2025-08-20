package main

import (
	"flag"
	"log/slog"
	"os"

	"bulbul/internal/validation"
)

func main() {
	var baseURL string
	flag.StringVar(&baseURL, "url", "http://localhost:8081", "Base URL for API validation")
	flag.Parse()

	slog.Info("Starting API validation", "url", baseURL)
	
	validator := validation.NewSpecValidator(baseURL)
	if err := validator.ValidateAll(); err != nil {
		slog.Error("❌ Валидация не пройдена", "error", err)
		os.Exit(1)
	}
	
	slog.Info("✅ Валидация успешно пройдена!")
}

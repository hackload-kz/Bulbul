package main

import (
	"flag"
	"log"
	"os"

	"bulbul/internal/validation"
)

func main() {
	var baseURL string
	flag.StringVar(&baseURL, "url", "http://localhost:8081", "Base URL for API validation")
	flag.Parse()

	log.Printf("Starting API validation against: %s", baseURL)
	
	validator := validation.NewSpecValidator(baseURL)
	if err := validator.ValidateAll(); err != nil {
		log.Fatalf("❌ Валидация не пройдена: %v", err)
		os.Exit(1)
	}
	
	log.Println("✅ Валидация успешно пройдена!")
}

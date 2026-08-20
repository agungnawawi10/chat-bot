package main

import (
	"log"
	"net/http"
	"os"

	"chat-bot/handler"
	"chat-bot/service"

	"github.com/joho/godotenv"
)

func main() {

	// Load .env
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Cek API key
	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY tidak ditemukan")
	}

	log.Println("API key berhasil dibaca")

	// Buat Gemini service
	geminiService := service.NewGeminiService()

	// Buat chat handler
	chatHandler := handler.NewChatHandler(
		geminiService,
	)

	// Register route
	http.HandleFunc(
		"/chat",
		chatHandler.Chat,
	)

	log.Println(
		"Server running on http://localhost:8080",
	)

	log.Fatal(
		http.ListenAndServe(":8080", nil),
	)
}

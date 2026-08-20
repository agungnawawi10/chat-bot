package main

import (
	"log"
	"net/http"
	"os"

	"chat-bot/database"
	"chat-bot/handler"
	"chat-bot/repository"
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

	// connect database
	db := database.Connect()
	defer db.Close()

	database.Migrate(db)

	chatRepository := repository.NewChatRepository(db)

	// Buat Gemini service
	// geminiService := service.NewGeminiService()

	// Buat chat handler
	// chatHandler := handler.NewChatHandler(
	// 	geminiService,
	// 	chatRepository,
	// )

	mockService := service.NewMockGeminiService()

	chatHandler := handler.NewChatHandler(
		mockService,
		chatRepository,
	)

	// Register route
	http.HandleFunc(
		"/chat",
		chatHandler.Chat,
	)

	http.HandleFunc(
		"/conversations",
		chatHandler.GetConversations,
	)

	http.HandleFunc(
		"/conversations/",
		chatHandler.ConversationRouter,
	)

	log.Println(
		"Server running on http://localhost:8080",
	)

	log.Fatal(
		http.ListenAndServe(":8080", nil),
	)
}

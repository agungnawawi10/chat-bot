package main

import (
	"context"
	// "fmt"
	"log"
	"net/http"
	"os"
	// "fmt"

	"chat-bot/database"
	"chat-bot/handler"
	"chat-bot/rag"
	"chat-bot/repository"
	"chat-bot/service"

	"github.com/joho/godotenv"
)

func main() {

	// 1. Load .env

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 2. Cek API key

	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY tidak ditemukan")
	}

	log.Println("API key berhasil dibaca")

	// 3. Connect database

	db := database.Connect()
	defer db.Close()

	database.Migrate(db)

	chatRepository := repository.NewChatRepository(db)

	vectorRepositoryConfig := &repository.VectorRepositoryConfig{
		Host:       "localhost",
		Port:       6334,
		Collection: "documents",
	}

	vectorRepository, err := repository.NewVectorRepository(vectorRepositoryConfig)

	if err != nil {
		log.Fatal(
			"Failed to connect to Qdrant:",
			err,
		)
	}

	// 4. Buat services

	// geminiService := service.NewMockGeminiService()

	geminiService := service.NewGeminiService()

	embeddingService := service.NewEmbeddingService()

	// Load Documents
	
	ctx := context.Background()

	err = rag.IndexDocument(
		ctx,
		"documents/company.txt",
		"company",
		embeddingService,
		vectorRepository,
	)

	if err != nil {
		log.Fatal(err)
	}

	err = rag.IndexDocument(
		ctx,
		"documents/profile_agung.txt",
		"profile_agung",
		embeddingService,
		vectorRepository,
	)

	if err != nil {
		log.Fatal(err)
	}

	// 9. Buat RAG Service

	ragService := service.NewRAGService(
		embeddingService,
		vectorRepository,
		geminiService,
	)

	// 10. Buat Chat Handler

	chatHandler := handler.NewChatHandler(
		geminiService,
		ragService,
		chatRepository,
	)

	// 11. Register routes

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
	// 12. Start server

	log.Println(
		"Server running on http://localhost:8080",
	)

	log.Fatal(
		http.ListenAndServe(
			":8080",
			nil,
		),
	)
}

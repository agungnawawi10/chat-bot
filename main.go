package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"

	"chat-bot/database"
	"chat-bot/handler"
	"chat-bot/rag"
	"chat-bot/repository"
	"chat-bot/service"

	"github.com/joho/godotenv"
)

var tmpl *template.Template

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY tidak ditemukan")
	}

	log.Println("API key berhasil dibaca")

	db := database.Connect()
	defer db.Close()

	database.Migrate(db)

	chatRepository := repository.NewChatRepository(db)

	vectorRepositoryConfig := &repository.VectorRepositoryConfig{
		Host:       os.Getenv("QDRANT_HOST"),
		Port:       6334,
		Collection: os.Getenv("QDRANT_COLLECTION"),
	}

	if vectorRepositoryConfig.Host == "" {
		vectorRepositoryConfig.Host = "localhost"
	}

	if vectorRepositoryConfig.Collection == "" {
		vectorRepositoryConfig.Collection = "documents"
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Config loaded:")
	log.Println("  QDRANT_HOST:", vectorRepositoryConfig.Host)
	log.Println("  QDRANT_COLLECTION:", vectorRepositoryConfig.Collection)
	log.Println("  SERVER_PORT:", port)

	vectorRepository, err := repository.NewVectorRepository(vectorRepositoryConfig)
	if err != nil {
		log.Fatal("Failed to connect to Qdrant:", err)
	}

	geminiService := service.NewGeminiService()
	embeddingService := service.NewEmbeddingService()

	ctx := context.Background()

	err = rag.IndexDocument(ctx, "documents/company.txt", "company", embeddingService, vectorRepository)
	if err != nil {
		log.Fatal(err)
	}

	err = rag.IndexDocument(ctx, "documents/profile_agung.txt", "profile_agung", embeddingService, vectorRepository)
	if err != nil {
		log.Fatal(err)
	}

	ragService := service.NewRAGService(embeddingService, vectorRepository, geminiService)

	chatHandler := handler.NewChatHandler(geminiService, ragService, chatRepository)

	tmpl = template.Must(template.ParseFiles("templates/index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/api/chat", chatHandler.Chat)
	http.HandleFunc("/conversations", chatHandler.GetConversations)
	http.HandleFunc("/conversations/", chatHandler.ConversationRouter)

	log.Println("Server running on http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

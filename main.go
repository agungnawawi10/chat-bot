package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"fmt"

	"chat-bot/database"
	"chat-bot/handler"
	"chat-bot/rag"
	"chat-bot/repository"
	"chat-bot/service"

	"github.com/joho/godotenv"
)

func main() {

	// ========================================
	// 1. Load .env
	// ========================================

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// ========================================
	// 2. Cek API key
	// ========================================

	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY tidak ditemukan")
	}

	log.Println("API key berhasil dibaca")

	// ========================================
	// 3. Connect database
	// ========================================

	db := database.Connect()
	defer db.Close()

	database.Migrate(db)

	chatRepository := repository.NewChatRepository(db)

	// ========================================
	// 4. Buat services
	// ========================================

	geminiService := service.NewGeminiService()

	embeddingService := service.NewEmbeddingService()

	// ========================================
	// 5. Load document
	// ========================================

	document, err := rag.LoadDocument(
		"documents/company.txt",
	)

	if err != nil {
		log.Fatal(
			"Failed to load document:",
			err,
		)
	}

	// ========================================
	// 6. Chunking
	// ========================================

	chunks := rag.ChunkText(
		document,
		100,
		20,
	)

	log.Printf(
		"Total chunks: %d",
		len(chunks),
	)

	// ========================================
	// 7. Connect Qdrant
	// ========================================

	vectorRepository, err := repository.NewVectorRepository()

	if err != nil {
		log.Fatal(
			"Failed to connect to Qdrant:",
			err,
		)
	}

	// ========================================
	// 8. Index document ke Qdrant
	// ========================================

	ctx := context.Background()

	for i, chunk := range chunks {

		log.Printf(
			"Processing chunk %d/%d...",
			i+1,
			len(chunks),
		)

		// Generate embedding
		embedding, err := embeddingService.GenerateEmbedding(
			chunk,
		)

		if err != nil {
			log.Fatalf(
				"Failed to generate embedding for chunk %d: %v",
				i+1,
				err,
			)
		}

		// Simpan ke Qdrant
		err = vectorRepository.SaveChunk(
			ctx,
			fmt.Sprintf("chunk-%d", i+1),
			chunk,
			embedding,
			"company.txt",
			i,
		)

		if err != nil {
			log.Fatalf(
				"Failed to save chunk %d: %v",
				i+1,
				err,
			)
		}

		log.Printf(
			"Chunk %d berhasil disimpan ke Qdrant",
			i+1,
		)
	}

	// ========================================
	// 9. Buat RAG Service
	// ========================================

	ragService := service.NewRAGService(
		embeddingService,
		vectorRepository,
		geminiService,
	)

	// ========================================
	// 10. Buat Chat Handler
	// ========================================

	chatHandler := handler.NewChatHandler(
		geminiService,
		ragService,
		chatRepository,
	)

	// ========================================
	// 11. Register routes
	// ========================================

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

	// ========================================
	// 12. Start server
	// ========================================

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

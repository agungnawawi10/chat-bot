package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"chat-bot/model"
	"chat-bot/repository"
	"chat-bot/service"

	"github.com/google/uuid"
)

// type ChatHandler struct {
// 	GeminiService  *service.GeminiService
// 	ChatRepository *repository.ChatRepository
// }

type ChatHandler struct {
	LLMService     service.LLMService
	RAGService     *service.RAGService
	ChatRepository *repository.ChatRepository
}

// func NewChatHandler(
// 	geminiService *service.GeminiService,
// 	chatRepository *repository.ChatRepository,
// ) *ChatHandler {
// 	return &ChatHandler{
// 		GeminiService:  geminiService,
// 		ChatRepository: chatRepository,
// 	}
// }

func NewChatHandler(
	llmService service.LLMService,
	ragService *service.RAGService,
	chatRepository *repository.ChatRepository,
) *ChatHandler {

	return &ChatHandler{
		LLMService:     llmService,
		RAGService:     ragService,
		ChatRepository: chatRepository,
	}
}

func (h *ChatHandler) Chat(w http.ResponseWriter, r *http.Request) {

	// 1. Cek HTTP method

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	// 2. Decode request

	var request model.ChatRequest

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		http.Error(
			w,
			"Invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	// 3. Validasi message

	if request.Message == "" {
		http.Error(
			w,
			"Message is required",
			http.StatusBadRequest,
		)
		return
	}

	// 4. Buat conversation baru

	if request.ConversationID == "" {

		request.ConversationID = uuid.New().String()

		err := h.ChatRepository.CreateConversation(
			request.ConversationID,
		)

		if err != nil {
			http.Error(
				w,
				"Failed to create conversation",
				http.StatusInternalServerError,
			)
			return
		}

		fmt.Println(
			"New Conversation ID:",
			request.ConversationID,
		)

	} else {

		// 5. Kalau ID sudah diberikan, cek database

		exists, err := h.ChatRepository.ConversationExists(
			request.ConversationID,
		)

		if err != nil {
			http.Error(
				w,
				"Failed to check conversation",
				http.StatusInternalServerError,
			)
			return
		}

		if !exists {
			http.Error(
				w,
				"Conversation not found",
				http.StatusNotFound,
			)
			return
		}
	}

	// 6. Simpan pesan user ke database

	err = h.ChatRepository.SaveMessage(
		request.ConversationID,
		"user",
		request.Message,
	)

	if err != nil {
		http.Error(
			w,
			"Failed to save user message",
			http.StatusInternalServerError,
		)
		return
	}

	// 7. Ambil history conversation

	conversation, err := h.ChatRepository.GetMessages(
		request.ConversationID,
	)

	if err != nil {
		http.Error(
			w,
			"Failed to get conversation messages",
			http.StatusInternalServerError,
		)
		return
	}

	// 8. Kirim history ke Gemini

	// answer, err := h.GeminiService.GenerateResponse(
	// 	conversation,
	// )

	answer, err := h.LLMService.GenerateResponse(
		conversation,
	)

	if err != nil {
		http.Error(
			w,
			"Failed to generate response: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// 9. Simpan jawaban Gemini

	err = h.ChatRepository.SaveMessage(
		request.ConversationID,
		"model",
		answer,
	)

	if err != nil {
		http.Error(
			w,
			"Failed to save model message",
			http.StatusInternalServerError,
		)
		return
	}

	// 10. Response ke client

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		model.ChatResponse{
			ConversationID: request.ConversationID,
			Answer:         answer,
		},
	)
}

func (h *ChatHandler) GetMessages(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodGet {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	// Ambil ID dari URL
	path := strings.TrimPrefix(
		r.URL.Path,
		"/conversations/",
	)

	// Contoh:
	// abc-123/messages

	parts := strings.Split(
		path,
		"/",
	)

	if len(parts) != 2 || parts[1] != "messages" {
		http.Error(
			w,
			"Invalid URL",
			http.StatusBadRequest,
		)
		return
	}

	conversationID := parts[0]

	exists, err := h.ChatRepository.ConversationExists(
		conversationID,
	)

	if err != nil {
		http.Error(
			w,
			"Failed to check conversation",
			http.StatusInternalServerError,
		)
		return
	}

	if !exists {
		http.Error(
			w,
			"Conversation not found",
			http.StatusNotFound,
		)
		return
	}

	messages, err := h.ChatRepository.GetMessages(
		conversationID,
	)

	if err != nil {
		http.Error(
			w,
			"Failed to get messages",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		messages,
	)
}

func (h *ChatHandler) ConversationRouter(
	w http.ResponseWriter,
	r *http.Request,
) {

	path := strings.TrimPrefix(
		r.URL.Path,
		"/conversations/",
	)

	parts := strings.Split(
		strings.Trim(path, "/"),
		"/",
	)

	if len(parts) == 1 {

		h.GetConversation(w, r)

		return
	}

	if len(parts) == 2 && parts[1] == "messages" {

		h.GetMessages(w, r)

		return
	}

	http.Error(
		w,
		"Invalid conversation endpoint",
		http.StatusNotFound,
	)
}

func (h *ChatHandler) GetConversations(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodGet {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	conversations, err := h.ChatRepository.GetConversations()

	if err != nil {
		http.Error(
			w,
			"Failed to get conversations",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		conversations,
	)
}

func (h *ChatHandler) GetConversation(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodGet {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	path := strings.TrimPrefix(
		r.URL.Path,
		"/conversations/",
	)

	conversationID := strings.TrimSuffix(
		path,
		"/",
	)

	if conversationID == "" {
		http.Error(
			w,
			"Conversation ID is required",
			http.StatusBadRequest,
		)
		return
	}

	conversation, err := h.ChatRepository.GetConversation(
		conversationID,
	)

	if err != nil {
		http.Error(
			w,
			"Conversation not found",
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		conversation,
	)
}

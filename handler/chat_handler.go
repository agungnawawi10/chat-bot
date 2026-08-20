package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"chat-bot/model"
	"chat-bot/service"

	"github.com/google/uuid"
)

type ChatHandler struct {
	GeminiService *service.GeminiService
	Conversations map[string][]model.GeminiContent
}

func NewChatHandler(geminiService *service.GeminiService) *ChatHandler {
	return &ChatHandler{
		GeminiService: geminiService,
		Conversations: make(map[string][]model.GeminiContent),
	}
}

func (h *ChatHandler) Chat(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

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

	if request.Message == "" {
		http.Error(
			w,
			"Message is required",
			http.StatusBadRequest,
		)
		return
	}

	if request.ConversationID == "" {
		request.ConversationID = uuid.New().String()
	}
	fmt.Println("Conversation ID:", request.ConversationID)

	conversation := h.Conversations[request.ConversationID]

	conversation = append(
		conversation,
		model.GeminiContent{
			Role: "user",
			Parts: []model.GeminiPart{
				{
					Text: request.Message,
				},
			},
		},
	)

	answer, err := h.GeminiService.GenerateResponse(
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

	conversation = append(
		conversation,
		model.GeminiContent{
			Role: "model",
			Parts: []model.GeminiPart{
				{
					Text: answer,
				},
			},
		},
	)

	h.Conversations[request.ConversationID] = conversation

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(
		model.ChatResponse{
			ConversationID: request.ConversationID,
			Answer:         answer,
		},
	)
}

package handler

import (
	"encoding/json"
	"net/http"

	"chat-bot/model"
	"chat-bot/service"
)

type ChatHandler struct {
	GeminiService *service.GeminiService
	Conversation  []model.GeminiContent
}

func NewChatHandler(geminiService *service.GeminiService) *ChatHandler {
	return &ChatHandler{
		GeminiService: geminiService,
		Conversation:  []model.GeminiContent{},
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

	// Simpan pesan user
	h.Conversation = append(
		h.Conversation,
		model.GeminiContent{
			Role: "user",
			Parts: []model.GeminiPart{
				{
					Text: request.Message,
				},
			},
		},
	)

	// Kirim conversation ke Gemini
	answer, err := h.GeminiService.GenerateResponse(
		h.Conversation,
	)

	if err != nil {
		http.Error(
			w,
			"Failed to generate response: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// Simpan jawaban Gemini
	h.Conversation = append(
		h.Conversation,
		model.GeminiContent{
			Role: "model",
			Parts: []model.GeminiPart{
				{
					Text: answer,
				},
			},
		},
	)

	// Response ke user
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(
		model.ChatResponse{
			Answer: answer,
		},
	)
}
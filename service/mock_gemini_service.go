package service

import "chat-bot/model"

type MockGeminiService struct{}

func NewMockGeminiService() *MockGeminiService {
	return &MockGeminiService{}
}

func (s *MockGeminiService) GenerateResponse(
	conversation []model.GeminiContent,
) (string, error) {

	if len(conversation) == 0 {
		return "Halo, belum ada percakapan.", nil
	}

	lastMessage := conversation[len(conversation)-1]

	userMessage := ""

	if len(lastMessage.Parts) > 0 {
		userMessage = lastMessage.Parts[0].Text
	}

	answer := "Mock response: kamu mengatakan \"" +
		userMessage +
		"\""

	return answer, nil
}
package service

import "chat-bot/model"

type LLMService interface {
	GenerateResponse(
		conversation []model.GeminiContent,
	) (string, error)
}

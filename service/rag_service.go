package service

import (
	"context"
	"fmt"
	"strings"

	"chat-bot/model"
	"chat-bot/repository"
)

type RAGService struct {
	EmbeddingService *EmbeddingService
	VectorRepository *repository.VectorRepository
	GeminiService    *GeminiService
}

func NewRAGService(
	embeddingService *EmbeddingService,
	vectorRepository *repository.VectorRepository,
	geminiService *GeminiService,
) *RAGService {
	return &RAGService{
		EmbeddingService: embeddingService,
		VectorRepository: vectorRepository,
		GeminiService:    geminiService,
	}
}

func (s *RAGService) GenerateAnswer(
	ctx context.Context,
	question string,
) (string, error) {

	// 1. Buat embedding dari pertanyaan
	queryEmbedding, err := s.EmbeddingService.GenerateEmbedding(question)

	if err != nil {
		return "", fmt.Errorf(
			"failed to generate query embedding: %w",
			err,
		)
	}

	// 2. Cari chunk yang paling relevan di Qdrant
	results, err := s.VectorRepository.SearchSimilar(
		ctx,
		queryEmbedding,
		3,
	)

	if err != nil {
		return "", fmt.Errorf(
			"failed to search vector database: %w",
			err,
		)
	}

	// 3. Buat context dari hasil pencarian
	var contextParts []string

	for _, result := range results {

		if result.Payload == nil {
			continue
		}

		contentValue, exists := result.Payload["content"]

		if !exists {
			continue
		}

		content := contentValue.GetStringValue()

		if content == "" {
			continue
		}

		contextParts = append(
			contextParts,
			content,
		)
	}

	if len(contextParts) == 0 {
		return "", fmt.Errorf(
			"no relevant context found",
		)
	}

	contextText := strings.Join(
		contextParts,
		"\n\n",
	)

	// 4. Buat prompt untuk Gemini
	prompt := fmt.Sprintf(
		`Jawab pertanyaan berdasarkan informasi berikut.

Context:
%s

Pertanyaan:
%s

Instruksi:
- Jawab berdasarkan context.
- Jangan mengarang informasi yang tidak terdapat di context.
- Jika jawabannya tidak tersedia dalam context, katakan bahwa informasi tersebut tidak tersedia.`,
		contextText,
		question,
	)

	// 5. Kirim context + question ke Gemini
	answer, err := s.GeminiService.GenerateResponse(
		[]model.GeminiContent{
			{
				Role: "user",
				Parts: []model.GeminiPart{
					{
						Text: prompt,
					},
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf(
			"failed to generate answer: %w",
			err,
		)
	}

	return answer, nil
}

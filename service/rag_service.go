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
	VectorRepository repository.VectorRepositoryInterface
	GeminiService    LLMService
	// GeminiService    *GeminiService
}

func NewRAGService(
	embeddingService *EmbeddingService,
	vectorRepository repository.VectorRepositoryInterface,
	geminiService LLMService,
	// geminiService *GeminiServic
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
	prompt := fmt.Sprintf(`
You are a RAG question-answering assistant.

Answer the user's question using ONLY the information provided in the CONTEXT.

STRICT RULES:
1. Use only information explicitly stated in the CONTEXT.
2. Do not use your own knowledge or outside information.
3. Do not invent, assume, or add any facts.
4. Do not mention services, features, programs, URLs, people, or other information
   that does not exist in the CONTEXT.
5. If the answer cannot be found in the CONTEXT, say:
   "Informasi tersebut tidak tersedia dalam dokumen."
6. Answer the question directly and concisely.
7. Preserve the exact meaning of the information in the CONTEXT.
8. Do not reinterpret, combine, or modify facts from different sentences.
9. Do not change one type of information into another.
10. When possible, use the wording from the CONTEXT directly.

CONTEXT:
%s

QUESTION:
%s

ANSWER:
`, contextText, question)

	// 5. Kirim context + question ke Gemini

	fmt.Println("\n=== CONTEXT SENT TO GEMINI ===")
	fmt.Println(contextText)
	fmt.Println("=== END CONTEXT ===")
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

package service

import (
	"context"
	"fmt"
	"strings"

	"chat-bot/model"
	"chat-bot/repository"
)

type RAGService struct {
	EmbeddingService EmbeddingServiceInterface
	VectorRepository repository.VectorRepositoryInterface
	GeminiService    GeminiService
}

func NewRAGService(
	embeddingService EmbeddingServiceInterface,
	vectorRepository repository.VectorRepositoryInterface,
	geminiService *GeminiService,
) *RAGService {
	return &RAGService{
		EmbeddingService: embeddingService,
		VectorRepository: vectorRepository,
		GeminiService:    *geminiService,
	}
}

func (s *RAGService) GenerateAnswer(
	ctx context.Context,
	question string,
) (string, error) {

	queryEmbedding, err := s.EmbeddingService.GenerateEmbedding(question)

	if err != nil {
		return "", fmt.Errorf(
			"failed to generate query embedding: %w",
			err,
		)
	}

	results, err := s.VectorRepository.SearchSimilar(
		ctx,
		queryEmbedding,
		3,
	)

	fmt.Println("\n=== RETRIEVAL DEBUG ===")

	for i, result := range results {
		fmt.Printf("\nResult %d\n", i+1)
		fmt.Println("Score:", result.Score)

		if result.Payload != nil {
			fmt.Println("Document:", result.Payload["document_name"])
			fmt.Println("Chunk Index:", result.Payload["chunk_index"])
			fmt.Println("Content:", result.Payload["content"])
		}
	}

	fmt.Println("=== END RETRIEVAL DEBUG ===")

	if err != nil {
		return "", fmt.Errorf(
			"failed to search vector database: %w",
			err,
		)
	}

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

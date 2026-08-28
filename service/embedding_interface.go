package service

type EmbeddingServiceInterface interface {
	GenerateEmbedding(text string) ([]float32, error)
}

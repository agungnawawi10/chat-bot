package repository

import (
	"context"

	"chat-bot/model"

	"github.com/qdrant/go-client/qdrant"
)

type ChatRepositoryInterface interface {
	CreateConversation(ctx context.Context, id string) error
	ConversationExists(ctx context.Context, id string) (bool, error)
	SaveMessage(ctx context.Context, conversationID string, role string, content string) error
	GetMessages(ctx context.Context, conversationID string) ([]model.GeminiContent, error)
	GetConversations(ctx context.Context) ([]model.ConversationResponse, error)
	GetConversation(ctx context.Context, id string) (*model.Conversation, error)
	UpdateConversationTimestamp(ctx context.Context, id string) error
}

type VectorRepositoryInterface interface {
	SaveChunk(
		ctx context.Context,
		id string,
		content string,
		embedding []float32,
		documentName string,
		chunkIndex int,
	) error
	SearchSimilar(
		ctx context.Context,
		embedding []float32,
		limit uint64,
	) ([]*qdrant.ScoredPoint, error)
}

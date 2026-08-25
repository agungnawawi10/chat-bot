package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"chat-bot/model"

	"github.com/qdrant/go-client/qdrant"
)

type MockChatRepository struct {
	mu            sync.Mutex
	conversations map[string]*model.Conversation
	messages      map[string][]model.GeminiContent
}

func NewMockChatRepository() *MockChatRepository {
	return &MockChatRepository{
		conversations: make(map[string]*model.Conversation),
		messages:      make(map[string][]model.GeminiContent),
	}
}

func (r *MockChatRepository) CreateConversation(
	ctx context.Context,
	id string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.conversations[id]; exists {
		return fmt.Errorf("conversation with ID %s already exists", id)
	}

	now := time.Now()
	r.conversations[id] = &model.Conversation{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return nil
}

func (r *MockChatRepository) ConversationExists(
	ctx context.Context,
	id string,
) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.conversations[id]
	return exists, nil
}

func (r *MockChatRepository) SaveMessage(
	ctx context.Context,
	conversationID string,
	role string,
	content string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conversation, exists := r.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation %s not found", conversationID)
	}

	now := time.Now()
	conversation.UpdatedAt = now

	msg := model.GeminiContent{
		Role: role,
		Parts: []model.GeminiPart{
			{
				Text: content,
			},
		},
	}

	r.messages[conversationID] = append(
		r.messages[conversationID],
		msg,
	)
	return nil
}

func (r *MockChatRepository) GetMessages(
	ctx context.Context,
	conversationID string,
) ([]model.GeminiContent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	messages, exists := r.messages[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation %s not found", conversationID)
	}

	return messages, nil
}

func (r *MockChatRepository) GetConversations(
	ctx context.Context,
) ([]model.ConversationResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var conversations []model.ConversationResponse
	for id := range r.conversations {
		conversations = append(
			conversations,
			model.ConversationResponse{ID: id},
		)
	}
	return conversations, nil
}

func (r *MockChatRepository) GetConversation(
	ctx context.Context,
	id string,
) (*model.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	conversation, exists := r.conversations[id]
	if !exists {
		return nil, fmt.Errorf("conversation %s not found", id)
	}
	return conversation, nil
}

func (r *MockChatRepository) UpdateConversationTimestamp(
	ctx context.Context,
	id string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conversation, exists := r.conversations[id]
	if !exists {
		return fmt.Errorf("conversation %s not found", id)
	}

	now := time.Now()
	conversation.UpdatedAt = now
	return nil
}

type MockVectorRepository struct {
	mu       sync.Mutex
	chunks   map[string]*MockChunk
}

type MockChunk struct {
	ID           string
	Content      string
	Embedding    []float32
	DocumentName string
	ChunkIndex   int
}

func NewMockVectorRepository() *MockVectorRepository {
	return &MockVectorRepository{
		chunks: make(map[string]*MockChunk),
	}
}

func (r *MockVectorRepository) SaveChunk(
	ctx context.Context,
	id string,
	content string,
	embedding []float32,
	documentName string,
	chunkIndex int,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.chunks[id] = &MockChunk{
		ID:           id,
		Content:      content,
		Embedding:    embedding,
		DocumentName: documentName,
		ChunkIndex:   chunkIndex,
	}
	return nil
}

func (r *MockVectorRepository) SearchSimilar(
	ctx context.Context,
	embedding []float32,
	limit uint64,
) ([]*qdrant.ScoredPoint, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var results []*qdrant.ScoredPoint
	for _, chunk := range r.chunks {
		score := computeSimilarity(embedding, chunk.Embedding)
		results = append(results, &qdrant.ScoredPoint{
			Id:      qdrant.NewIDNum(uint64(chunk.ChunkIndex)),
			Score:   score,
			Payload: chunk.ToPayload(),
		})
	}

	if uint64(len(results)) > limit {
		results = results[:limit]
	}
	return results, nil
}

func computeSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (normA * normB)
}

func (c *MockChunk) ToPayload() map[string]*qdrant.Value {
	payload := make(map[string]*qdrant.Value)
	
	v, _ := qdrant.NewValue(c.ID)
	payload["id"] = v
	
	v, _ = qdrant.NewValue(c.Content)
	payload["content"] = v
	
	v, _ = qdrant.NewValue(c.DocumentName)
	payload["document_name"] = v
	
	v, _ = qdrant.NewValue(int64(c.ChunkIndex))
	payload["chunk_index"] = v
	
	return payload
}

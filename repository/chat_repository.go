package repository

import (
	"database/sql"
	"fmt"

	"chat-bot/model"
)

type ChatRepository struct {
	DB *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{
		DB: db,
	}
}

func (r *ChatRepository) CreateConversation(id string) error {

	query := `
		INSERT INTO conversations (id)
		VALUES (?)
	`

	_, err := r.DB.Exec(query, id)

	// untuk cek apakah ada id yang sama yang sudah tersimpan di database
	if err != nil {
		if sqliteErr, ok := err.(interface{ Code() int32 }); ok {
			if sqliteErr.Code() == 19 {
				return fmt.Errorf(
					"conversation with ID %s already exists", id,
				)
			}
		}
		return fmt.Errorf(
			"failed to create conversation: %w",
			err,
		)
	}

	return nil
}

func (r *ChatRepository) ConversationExists(
	id string,
) (bool, error) {

	query := `
		SELECT EXISTS(
			SELECT 1
			FROM conversations
			WHERE id = ?
		)
	`

	var exists bool

	err := r.DB.QueryRow(
		query,
		id,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf(
			"failed to check conversation: %w",
			err,
		)
	}

	return exists, nil
}

func (r *ChatRepository) SaveMessage(
	conversationID string,
	role string,
	content string,
) error {

	query := `
		INSERT INTO messages (
			conversation_id,
			role,
			content
		)
		VALUES (?, ?, ?)
	`

	_, err := r.DB.Exec(
		query,
		conversationID,
		role,
		content,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to save message: %w",
			err,
		)
	}

	return nil
}

func (r *ChatRepository) GetMessages(
	conversationID string,
) ([]model.GeminiContent, error) {

	query := `
		SELECT role, content
		FROM messages
		WHERE conversation_id = ?
		ORDER BY id ASC
	`

	rows, err := r.DB.Query(
		query,
		conversationID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get messages: %w",
			err,
		)
	}

	defer rows.Close()

	var messages []model.GeminiContent

	for rows.Next() {

		var role string
		var content string

		err := rows.Scan(
			&role,
			&content,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to scan message: %w",
				err,
			)
		}

		messages = append(
			messages,
			model.GeminiContent{
				Role: role,
				Parts: []model.GeminiPart{
					{
						Text: content,
					},
				},
			},
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed while reading messages: %w",
			err,
		)
	}

	return messages, nil
}

func (r *ChatRepository) GetConversations() (
	[]model.ConversationResponse,
	error,
) {

	query := `
		SELECT id
		FROM conversations
		ORDER BY id DESC
	`

	rows, err := r.DB.Query(query)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get conversations: %w",
			err,
		)
	}

	defer rows.Close()

	var conversations []model.ConversationResponse

	for rows.Next() {

		var conversation model.ConversationResponse

		err := rows.Scan(
			&conversation.ID,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to scan conversation: %w",
				err,
			)
		}

		conversations = append(
			conversations,
			conversation,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *ChatRepository) GetConversation(
	id string,
) (*model.Conversation, error) {

	query := `
		SELECT id, created_at, updated_at
		FROM conversations
		WHERE id = ?
	`

	var conversation model.Conversation

	err := r.DB.QueryRow(
		query,
		id,
	).Scan(
		&conversation.ID,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get conversation: %w",
			err,
		)
	}

	return &conversation, nil
}

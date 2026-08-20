package model

type APIChatRequest struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
}

type APIChatResponse struct {
	ConversationID string `json:"conversation_id"`
	Answer         string `json:"answer"`
}

type APIConversationResponse struct {
	ID string `json:"id"`
}

type APIMessageResponse struct {
	ID             int64  `json:"id"`
	ConversationID string `json:"conversation_id"`
	Role           string `json:"role"`
	Content        string `json:"content"`
}

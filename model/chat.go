package model

type ChatRequest struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
}

type ChatResponse struct {
	ConversationID string `json:"conversation_id"`
	Answer         string `json:"answer"`
}

// Request ke Gemini
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

// Response dari Gemini
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

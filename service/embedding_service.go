package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type EmbeddingService struct {
	APIKey string
}

type EmbeddingRequest struct {
	Model   string           `json:"model"`
	Content EmbeddingContent `json:"content"`
}

type EmbeddingContent struct {
	Parts []EmbeddingPart `json:"parts"`
}

type EmbeddingPart struct {
	Text string `json:"text"`
}

type EmbeddingResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

func NewEmbeddingService() *EmbeddingService {
	return &EmbeddingService{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	}
}

func (s *EmbeddingService) GenerateEmbedding(
	text string,
) ([]float32, error) {

	requestData := EmbeddingRequest{
		Model: "models/gemini-embedding-001",
		Content: EmbeddingContent{
			Parts: []EmbeddingPart{
				{
					Text: text,
				},
			},
		},
	}

	requestBody, err := json.Marshal(requestData)

	if err != nil {
		return nil, err
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent?key=" + s.APIKey

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(requestBody),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"Gemini Embedding API error: %s",
			string(body),
		)
	}

	var response EmbeddingResponse

	err = json.Unmarshal(
		body,
		&response,
	)

	if err != nil {
		return nil, err
	}

	return response.Embedding.Values, nil
}

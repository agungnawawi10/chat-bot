package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"chat-bot/model"
)

type GeminiService struct {
	APIKey string
}

func NewGeminiService() *GeminiService {
	return &GeminiService{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	}
}

func (s *GeminiService) GenerateResponse(
	conversation []model.GeminiContent,
) (string, error) {

	geminiRequest := model.GeminiRequest{
		Contents: conversation,
	}

	requestBody, err := json.Marshal(geminiRequest)
	if err != nil {
		return "", err
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-3.6-flash:generateContent?key=" + s.APIKey

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(requestBody),
	)

	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Println("Gemini Status:", resp.Status)
	fmt.Println("Gemini Response:", string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Gemini API error: %s", resp.Status)
	}

	var geminiResponse model.GeminiResponse

	err = json.Unmarshal(body, &geminiResponse)
	if err != nil {
		return "", err
	}

	if len(geminiResponse.Candidates) == 0 {
		return "", fmt.Errorf("Gemini returned no candidates")
	}

	if len(geminiResponse.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("Gemini returned no content")
	}

	answer := geminiResponse.Candidates[0].Content.Parts[0].Text

	return answer, nil
}

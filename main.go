package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Answer string `json:"answer"`
}

// Request ke Gemini
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
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

func chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var chatRequest ChatRequest

	err := json.NewDecoder(r.Body).Decode(&chatRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if chatRequest.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	// Ambil API key dari .env
	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		http.Error(w, "GEMINI_API_KEY not found", http.StatusInternalServerError)
		return
	}

	// Request yang akan dikirim ke Gemini
	geminiRequest := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{
						Text: chatRequest.Message,
					},
				},
			},
		},
	}

	requestBody, err := json.Marshal(geminiRequest)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Endpoint Gemini
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-3.6-flash:generateContent?key=" + apiKey

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(requestBody),
	)

	if err != nil {
		http.Error(w, "Failed to create Gemini request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Kirim request ke Gemini
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to connect to Gemini", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	// Baca response Gemini
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read Gemini response", http.StatusInternalServerError)
		return
	}

	fmt.Println("Gemini Status:", resp.Status)
	fmt.Println("Gemini Response:", string(body))

	// Kalau Gemini error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		http.Error(
			w,
			"Gemini API error: "+resp.Status,
			resp.StatusCode,
		)
		return
	}

	var geminiResponse GeminiResponse

	err = json.Unmarshal(body, &geminiResponse)
	if err != nil {
		http.Error(w, "Failed to parse Gemini response", http.StatusInternalServerError)
		return
	}

	// Ambil jawaban dari Gemini
	answer := ""

	if len(geminiResponse.Candidates) > 0 &&
		len(geminiResponse.Candidates[0].Content.Parts) > 0 {

		answer = geminiResponse.Candidates[0].Content.Parts[0].Text
	}

	// Kirim jawaban ke user
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(ChatResponse{
		Answer: answer,
	})
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY tidak ditemukan")
	}

	log.Println("API key berhasil dibaca")

	http.HandleFunc("/chat", chatHandler)

	log.Println("Server running on http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

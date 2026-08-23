# AGENTS.md

## Quick Start

```bash
# Run the chatbot server
go run main.go

# Server runs on http://localhost:8080
```

## Architecture

- **Entry point**: `main.go` - loads `.env`, connects SQLite, runs HTTP server on port 8080
- **Package structure**:
  - `model/` - data structures (Gemini API models, chat request/response, database models)
  - `service/` - business logic (GeminiService, MockGeminiService, EmbeddingService, LLMService interface)
  - `handler/` - HTTP handlers (`/chat`, `/conversations`, `/conversations/:id`, `/conversations/:id/messages`)
  - `repository/` - database access layer
  - `database/` - SQLite connection and migrations
  - `rag/` - document loading and text chunking

## Key Conventions

- `.env` required: `GEMINI_API_KEY` (currently hardcoded, do not expose)
- Uses `modernc.org/sqlite` - no external SQLite binary needed
- Mock Gemini service enabled in `main.go:54` by default; uncomment lines 45-49 to use real Gemini API
- Gemini API models: `gemini-3.6-flash` for chat, `gemini-embedding-001` for embeddings
- Document to load: `documents/company.txt` (company info)
- Chunking defaults: `chunkSize=100, overlap=20`

## API Endpoints

- `POST /chat` - Send message with optional `conversation_id`, returns `conversation_id` + `answer`
- `GET /conversations` - List all conversation IDs
- `GET /conversations/:id` - Get conversation metadata
- `GET /conversations/:id/messages` - Get messages in a conversation

## Database

- SQLite file: `chatbot.db` (created in working directory)
- Tables: `conversations` (id, created_at, updated_at), `messages` (id, conversation_id, role, content, created_at)
- Migration runs automatically on startup via `database.Migrate()`

## Testing

- Use `MockGeminiService` for local testing (no API key required)
- Real Gemini integration requires valid `GEMINI_API_KEY` in `.env`

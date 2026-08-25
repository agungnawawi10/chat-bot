# Chat Bot AI

Chatbot berbasis Golang dengan Gemini API, PostgreSQL, dan Qdrant sebagai vector database.

## Requirements

Pastikan sudah menginstall:

- Go
- PostgreSQL
- Docker
- Git

Cek instalasi:

```bash
go version
psql --version
docker --version
git --version
```
### Clone Repository

```bash
git clone <repository-url>
cd chat-bot
```

### Install Dependencies

```bash 
go mod tidy
```

### Setup PostgreSQL

```bash
CREATE DATABASE chatbot;
```

### Setup Qdrant Vektor Database

```bash 
docker run -d --name qdrant -p 6333:6333 -p 6334:6334 qdrant/qdrant

docker ps
```
### Run Application

```bash 
go run .

```
package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

func Connect() *sql.DB {
	db, err := sql.Open("sqlite", "./chatbot.db")
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Database connected")

	return db
}

// migrate table chat-bot
func Migrate(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS conversations (
		id TEXT PRIMARY KEY,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conversation_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

		FOREIGN KEY (conversation_id)
			REFERENCES conversations(id)
			ON DELETE CASCADE
	);
	`

	_, err := db.Exec(query)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed")
}

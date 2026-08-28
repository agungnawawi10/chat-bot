package rag

import (
	"context"
	"fmt"
	"log"

	"chat-bot/repository"
	"chat-bot/service"
)

func IndexDocument(
	ctx context.Context,
	filePath string,
	documentName string,
	embeddingService service.EmbeddingServiceInterface,
	vectorRepository *repository.VectorRepository,
) error {

	// 1. Load document
	document, err := LoadDocument(filePath)

	if err != nil {
		return fmt.Errorf(
			"failed to load document %s: %w",
			filePath,
			err,
		)
	}

	// 2. Chunking
	chunks := ChunkText(
		document,
		100,
		20,
	)

	log.Printf(
		"Document %s: %d chunks",
		documentName,
		len(chunks),
	)

	// 3. Generate embedding + save ke Qdrant
	for i, chunk := range chunks {

		log.Printf(
			"Processing %s chunk %d/%d...",
			documentName,
			i+1,
			len(chunks),
		)

		embedding, err := embeddingService.GenerateEmbedding(
			chunk,
		)

		if err != nil {
			return fmt.Errorf(
				"failed to generate embedding for chunk %d: %w",
				i+1,
				err,
			)
		}

		err = vectorRepository.SaveChunk(
			ctx,
			fmt.Sprintf(
				"%s-chunk-%d",
				documentName,
				i+1,
			),
			chunk,
			embedding,
			documentName,
			i,
		)

		if err != nil {
			return fmt.Errorf(
				"failed to save chunk %d: %w",
				i+1,
				err,
			)
		}

		log.Printf(
			"%s chunk %d berhasil disimpan ke Qdrant",
			documentName,
			i+1,
		)
	}

	return nil
}

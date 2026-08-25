package repository

import (
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
)

type VectorRepository struct {
	Client     *qdrant.Client
	Collection string
}

func NewVectorRepository() (*VectorRepository, error) {

	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})

	if err != nil {
		return nil, err
	}

	return &VectorRepository{
		Client:     client,
		Collection: "documents",
	}, nil
}

func (r *VectorRepository) SaveChunk(
	ctx context.Context,
	id string,
	content string,
	embedding []float32,
	documentName string,
	chunkIndex int,
) error {

	_, err := r.Client.Upsert(
		ctx,
		&qdrant.UpsertPoints{
			CollectionName: r.Collection,
			Points: []*qdrant.PointStruct{
				{
					Id: qdrant.NewIDNum(uint64(chunkIndex)),
					Vectors: qdrant.NewVectors(
						embedding...,
					),
					Payload: qdrant.NewValueMap(
						map[string]any{
							"id":            id,
							"content":       content,
							"document_name": documentName,
							"chunk_index":   chunkIndex,
						},
					),
				},
			},
		},
	)

	if err != nil {
		return fmt.Errorf(
			"failed to save chunk to qdrant: %w",
			err,
		)
	}

	return nil
}

func (r *VectorRepository) SearchSimilar(
	ctx context.Context,
	embedding []float32,
	limit uint64,
) ([]*qdrant.ScoredPoint, error) {

	result, err := r.Client.Query(
		ctx,
		&qdrant.QueryPoints{
			CollectionName: r.Collection,
			Query:          qdrant.NewQuery(embedding...),
			Limit:          qdrant.PtrOf(limit),
			WithPayload:    qdrant.NewWithPayload(true),
		},
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to search similar vectors: %w",
			err,
		)
	}

	return result, nil
}

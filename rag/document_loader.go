package rag

import (
	"fmt"
	"os"
)

func LoadDocument(path string) (string, error) {

	data, err := os.ReadFile(path)

	if err != nil {
		return "", fmt.Errorf(
			"failed to read document: %w",
			err,
		)
	}

	return string(data), nil
}
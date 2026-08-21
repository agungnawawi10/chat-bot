package rag

import "strings"

func ChunkText(
	text string,
	chunkSize int,
	overlap int,
) []string {

	text = strings.TrimSpace(text)

	if chunkSize <= 0 || len(text) == 0 {
		return []string{}
	}

	if overlap < 0 {
		overlap = 0
	}

	if overlap >= chunkSize {
		overlap = chunkSize / 4
	}

	var chunks []string

	start := 0

	for start < len(text) {

		end := start + chunkSize

		if end >= len(text) {
			end = len(text)
		} else {

			// Cari spasi terakhir sebelum batas chunk.
			splitPos := strings.LastIndexAny(
				text[start:end],
				" \t\n",
			)

			if splitPos != -1 {
				end = start + splitPos
			}
		}

		chunk := strings.TrimSpace(
			text[start:end],
		)

		if chunk != "" {
			chunks = append(
				chunks,
				chunk,
			)
		}

		if end >= len(text) {
			break
		}

		// Mundur untuk membuat overlap.
		nextStart := end - overlap

		if nextStart <= start {
			nextStart = end
		}

		start = nextStart
	}

	return chunks
}
package rag

import (
	"strings"
)

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

			splitPos := findOptimalSplitPoint(
				text,
				start,
				end,
				chunkSize,
			)

			if splitPos != -1 {
				end = splitPos
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

		nextStart := end - overlap

		if nextStart <= start {
			nextStart = end
		}

		start = nextStart
	}

	return chunks
}

func findOptimalSplitPoint(
	text string,
	start int,
	end int,
	chunkSize int,
) int {

	if end >= len(text) {
		return len(text)
	}

	for end < len(text) {
		if text[end] == ' ' || text[end] == '\t' || text[end] == '\n' {
			break
		}
		end++
		if end-start > chunkSize*2 {
			break
		}
	}

	if end >= len(text) {
		return len(text)
	}

	lastPeriod := strings.LastIndex(
		text[start:end],
		".",
	)

	if lastPeriod > chunkSize/2 {
		return start + lastPeriod + 1
	}

	lastSpace := strings.LastIndexAny(
		text[start:end],
		" \t\n",
	)

	if lastSpace > chunkSize/2 {
		return start + lastSpace + 1
	}

	return end
}
package main

func SplitByMaxBytes(input []string, maxBytes int) [][]string {
	var result [][]string
	var currentChunk []string
	var currentSize int

	for _, s := range input {
		sLen := len(s)
		if currentSize+sLen > maxBytes {
			// Push current chunk
			result = append(result, currentChunk)
			// Start new chunk
			currentChunk = []string{s}
			currentSize = sLen
		} else {
			currentChunk = append(currentChunk, s)
			currentSize += sLen
		}
	}

	// Add last chunk
	if len(currentChunk) > 0 {
		result = append(result, currentChunk)
	}

	return result
}

func MergeChunks(chunks [][]string) []string {
	var merged []string
	for _, chunk := range chunks {
		merged = append(merged, chunk...)
	}
	return merged
}

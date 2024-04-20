package main

func splitStringIntoChunks(s string, chunkSize int) []string {
	var chunks []string
	for chunkSize < len(s) {
		chunks = append(chunks, s[:chunkSize])
		s = s[chunkSize:]
	}
	chunks = append(chunks, s)
	return chunks
}

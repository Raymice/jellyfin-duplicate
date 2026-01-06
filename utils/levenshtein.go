package utils

import (
	"strings"
)

// LevenshteinDistance calculates the Levenshtein distance between two strings
// This is a pure Go implementation without external dependencies
func LevenshteinDistance(s1, s2 string) int {
	// Convert strings to runes for proper Unicode handling
	r1 := []rune(s1)
	r2 := []rune(s2)

	len1 := len(r1)
	len2 := len(r2)

	// Create a matrix to store distances
	distances := make([][]int, len1+1)
	for i := range distances {
		distances[i] = make([]int, len2+1)
	}

	// Initialize the matrix
	for i := 0; i <= len1; i++ {
		distances[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		distances[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}

			distances[i][j] = min(
				distances[i-1][j]+1,      // deletion
				distances[i][j-1]+1,      // insertion
				distances[i-1][j-1]+cost, // substitution
			)
		}
	}

	return distances[len1][len2]
}

// calculatePathSimilarity computes the similarity percentage between two paths
// using the Levenshtein distance algorithm implemented in pure Go
// Note: File extensions are excluded from the comparison
func CalculatePathSimilarity(path1, path2 string) int {
	// Remove file extensions before comparison
	path1WithoutExt := removeFileExtension(path1)
	path2WithoutExt := removeFileExtension(path2)

	// Implement Levenshtein distance algorithm
	distance := LevenshteinDistance(path1WithoutExt, path2WithoutExt)

	// Calculate maximum possible distance
	maxLen := len(path1WithoutExt)
	if len(path2WithoutExt) > maxLen {
		maxLen = len(path2WithoutExt)
	}

	if maxLen == 0 {
		return 100
	}

	// Calculate similarity percentage
	similarity := 100 - (distance * 100 / maxLen)
	return similarity
}

// removeFileExtension removes the file extension from a path
// Example: "/movies/movie.mkv" → "/movies/movie"
func removeFileExtension(path string) string {
	// Find the last dot in the path
	lastDotIndex := strings.LastIndex(path, ".")

	// If no dot found, or dot is at the start, return original path
	if lastDotIndex <= 0 {
		return path
	}

	// Check if the dot is part of a file extension
	// Look for common extension patterns
	lastSlashIndex := strings.LastIndex(path, "/")

	// If there's a slash after the last dot, it's not an extension
	if lastSlashIndex > lastDotIndex {
		return path
	}

	// Remove everything after the last dot (the extension)
	return path[:lastDotIndex]
}

// min returns the minimum of multiple integers
func min(values ...int) int {
	if len(values) == 0 {
		return 0
	}

	minVal := values[0]
	for _, val := range values[1:] {
		if val < minVal {
			minVal = val
		}
	}
	return minVal
}

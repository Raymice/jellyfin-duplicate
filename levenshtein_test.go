package main

import (
	"fmt"
	"strings"
	"testing"
)

// levenshteinDistance calculates the Levenshtein distance between two strings
// This is a pure Go implementation without external dependencies
func levenshteinDistance(s1, s2 string) int {
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
func calculatePathSimilarity(path1, path2 string) int {
	// Implement Levenshtein distance algorithm
	distance := levenshteinDistance(path1, path2)
	
	// Calculate maximum possible distance
	maxLen := len(path1)
	if len(path2) > maxLen {
		maxLen = len(path2)
	}
	
	if maxLen == 0 {
		return 100
	}
	
	// Calculate similarity percentage
	similarity := 100 - (distance * 100 / maxLen)
	return similarity
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

// Test cases for our Levenshtein distance implementation
func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{
			name:     "identical strings",
			s1:       "hello",
			s2:       "hello",
			expected: 0,
		},
		{
			name:     "empty strings",
			s1:       "",
			s2:       "",
			expected: 0,
		},
		{
			name:     "one empty string",
			s1:       "hello",
			s2:       "",
			expected: 5,
		},
		{
			name:     "single character difference",
			s1:       "kitten",
			s2:       "sitting",
			expected: 3,
		},
		{
			name:     "completely different",
			s1:       "abc",
			s2:       "xyz",
			expected: 3,
		},
		{
			name:     "unicode characters",
			s1:       "café",
			s2:       "cafe",
			expected: 1,
		},
		{
			name:     "file paths",
			s1:       "/movies/inception.mkv",
			s2:       "/movies/inception_2010.mkv",
			expected: 6,
		},
		{
			name:     "file paths without extensions",
			s1:       "/movies/inception",
			s2:       "/movies/inception_2010",
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := levenshteinDistance(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

func TestCalculatePathSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		path1    string
		path2    string
		expected int
	}{
		{
			name:     "identical paths",
			path1:    "/movies/movie.mkv",
			path2:    "/movies/movie.mkv",
			expected: 100,
		},
		{
			name:     "very similar paths",
			path1:    "/movies/inception.mkv",
			path2:    "/movies/inception_2010.mkv",
			expected: 88, // 100 - (6 * 100 / 21) ≈ 88
		},
		{
			name:     "completely different paths",
			path1:    "/movies/a.mkv",
			path2:    "/tv/show/s01e01.mkv",
			expected: 0, // Should be very low similarity
		},
		{
			name:     "empty paths",
			path1:    "",
			path2:    "",
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePathSimilarity(tt.path1, tt.path2)
			// Allow some tolerance for the similarity calculation
			tolerance := 2
			if result < tt.expected-tolerance || result > tt.expected+tolerance {
				t.Errorf("calculatePathSimilarity(%q, %q) = %d, want approximately %d", tt.path1, tt.path2, result, tt.expected)
			}
		})
	}
}

// Main function to demonstrate the Levenshtein distance calculation
func main() {
	fmt.Println("Levenshtein Distance Examples:")
	fmt.Printf("Distance between 'kitten' and 'sitting': %d\n", levenshteinDistance("kitten", "sitting"))
	fmt.Printf("Distance between 'hello' and 'world': %d\n", levenshteinDistance("hello", "world"))
	fmt.Printf("Distance between 'abc' and 'abc': %d\n", levenshteinDistance("abc", "abc"))
	
	fmt.Println("\nPath Similarity Examples:")
	fmt.Printf("Similarity between identical paths: %d%%\n", calculatePathSimilarity("/movies/test.mkv", "/movies/test.mkv"))
	fmt.Printf("Similarity between similar paths: %d%%\n", calculatePathSimilarity("/movies/inception.mkv", "/movies/inception_2010.mkv"))
	fmt.Printf("Similarity between different paths: %d%%\n", calculatePathSimilarity("/movies/a.mkv", "/tv/show/episode.mkv"))
}
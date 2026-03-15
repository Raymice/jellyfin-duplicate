package utils_test

import (
	"jellyfin-duplicate/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name       string // description of this test case
		moviePath1 string
		moviePath2 string
		want       int
	}{
		{
			name:       "Not identical paths",
			moviePath1: "Largo winch 2 (2011).avi",
			moviePath2: "Largo Winch II (2011).mkv",
			want:       6,
		},
		{
			name:       "Identical paths",
			moviePath1: "Largo winch 2 (2011).avi",
			moviePath2: "Largo winch 2 (2011).avi",
			want:       0,
		},
		{
			name:       "Completely different paths",
			moviePath1: "Largo winch 2 (2011).avi",
			moviePath2: "Inception (2010).mkv",
			want:       16,
		},
		{
			name:       "Paths with different extensions",
			moviePath1: "Largo winch 2 (2011).avi",
			moviePath2: "Largo winch 2 (2011).mkv",
			want:       3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.LevenshteinDistance(tt.moviePath1, tt.moviePath2)
			assert.Equal(t, tt.want, got, "Should return the expected Levenshtein distance")
		})
	}
}

func TestCalculatePathSimilarity(t *testing.T) {
	tests := []struct {
		name       string // description of this test case
		moviePath1 string
		moviePath2 string
		want       int
	}{
		{
			name:       "Not identical paths",
			moviePath1: "Largo winch 2 (2011).avi",
			moviePath2: "Largo Winch II (2011).mkv",
			want:       86,
		},
		{
			name:       "Identical paths",
			moviePath1: "Largo winch 2 (2011).avi",
			moviePath2: "Largo winch 2 (2011).avi",
			want:       100,
		},
		{
			name:       "Completely different paths",
			moviePath1: "Largo winch 2 (2011).avi",
			moviePath2: "Inception (2010).mkv",
			want:       35,
		},
		{
			name:       "Paths with different extensions",
			moviePath1: "Largo winch 2 (2011).avi",
			moviePath2: "Largo winch 2 (2011).mkv",
			want:       100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.CalculatePathSimilarity(tt.moviePath1, tt.moviePath2)
			assert.Equal(t, tt.want, got, "Should return the expected similarity percentage")
		})
	}
}

func TestRemoveFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Simple filename with extension",
			path:     "movie.mkv",
			expected: "movie",
		},
		{
			name:     "Path with filename and extension",
			path:     "/movies/the-matrix.avi",
			expected: "/movies/the-matrix",
		},
		{
			name:     "Path with dots in name",
			path:     "/movies/my.movie.name.mkv",
			expected: "/movies/my.movie.name",
		},
		{
			name:     "Filename without extension",
			path:     "movie",
			expected: "movie",
		},
		{
			name:     "Path without extension",
			path:     "/some/path/movie",
			expected: "/some/path/movie",
		},
		{
			name:     "Empty string",
			path:     "",
			expected: "",
		},
		{
			name:     "Dot at start of filename",
			path:     ".hidden",
			expected: ".hidden",
		},
		{
			name:     "Double extension",
			path:     "archive.tar.gz",
			expected: "archive.tar",
		},
		{
			name:     "Windows path with extension",
			path:     "C:\\movies\\my-movie.mkv",
			expected: "C:\\movies\\my-movie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.RemoveFileExtension(tt.path)
			assert.Equal(t, tt.expected, got, "Should remove the file extension correctly")
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		values   []int
		expected int
	}{
		{
			name:     "Single value",
			values:   []int{5},
			expected: 5,
		},
		{
			name:     "Two values - first is min",
			values:   []int{3, 5},
			expected: 3,
		},
		{
			name:     "Two values - second is min",
			values:   []int{5, 3},
			expected: 3,
		},
		{
			name:     "Multiple values",
			values:   []int{10, 5, 15, 3, 8},
			expected: 3,
		},
		{
			name:     "Multiple values - zero is min",
			values:   []int{10, 5, 0, 3},
			expected: 0,
		},
		{
			name:     "Negative numbers",
			values:   []int{5, -10, 3, -5},
			expected: -10,
		},
		{
			name:     "All same values",
			values:   []int{5, 5, 5},
			expected: 5,
		},
		{
			name:     "Empty values",
			values:   []int{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.Min(tt.values...)
			assert.Equal(t, tt.expected, got, "Should return the minimum value")
		})
	}
}

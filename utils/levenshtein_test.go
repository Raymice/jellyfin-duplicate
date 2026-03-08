package utils_test

import (
	"jellyfin-duplicate/utils"
	"testing"
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
			if tt.want != got {
				t.Errorf("LevenshteinDistance() = %v, want %v", got, tt.want)
			}
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
			if tt.want != got {
				t.Errorf("CalculatePathSimilarity() = %v, want %v", got, tt.want)
			}
		})
	}
}

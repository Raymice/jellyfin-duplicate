package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUserPlayStatusDTO tests the UserPlayStatusDTO struct
func TestUserPlayStatusDTO(t *testing.T) {
	tests := []struct {
		name           string
		status         UserPlayStatusDTO
		expectedPlayed bool
	}{
		{
			name: "User played movie",
			status: UserPlayStatusDTO{
				UserID:    "user1",
				UserName:  "John",
				Played:    true,
				PlayCount: 5,
			},
			expectedPlayed: true,
		},
		{
			name: "User not played movie",
			status: UserPlayStatusDTO{
				UserID:    "user2",
				UserName:  "Jane",
				Played:    false,
				PlayCount: 0,
			},
			expectedPlayed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedPlayed, tt.status.Played)
			assert.NotEmpty(t, tt.status.UserID)
		})
	}
}

// TestPlayStatusDiscrepancyDTO tests the PlayStatusDiscrepancyDTO struct
func TestPlayStatusDiscrepancyDTO(t *testing.T) {
	tests := []struct {
		name        string
		discrepancy PlayStatusDiscrepancyDTO
	}{
		{
			name: "Discrepancy between two movies",
			discrepancy: PlayStatusDiscrepancyDTO{
				UserID:        "user1",
				UserName:      "John",
				MovieToUpdate: "movie-id-2",
				MovieName:     "Movie Title",
			},
		},
		{
			name: "Multiple discrepancies",
			discrepancy: PlayStatusDiscrepancyDTO{
				UserID:        "user2",
				UserName:      "Jane",
				MovieToUpdate: "movie-id-3",
				MovieName:     "Another Movie",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.discrepancy.UserID)
			assert.NotEmpty(t, tt.discrepancy.MovieToUpdate)
			assert.NotEmpty(t, tt.discrepancy.MovieName)
		})
	}
}

// TestDuplicateResultDTO tests the DuplicateResultDTO struct
func TestDuplicateResultDTO(t *testing.T) {
	tests := []struct {
		name              string
		duplicate         DuplicateResultDTO
		expectedDuplicate bool
	}{
		{
			name: "Confirmed duplicate",
			duplicate: DuplicateResultDTO{
				IsDuplicate:             true,
				Similarity:              98,
				HasIdenticalPlayStatus:  true,
				PlayStatusDiscrepancies: []PlayStatusDiscrepancyDTO{},
			},
			expectedDuplicate: true,
		},
		{
			name: "Similar but not identical",
			duplicate: DuplicateResultDTO{
				IsDuplicate:            false,
				Similarity:             75,
				HasIdenticalPlayStatus: false,
				PlayStatusDiscrepancies: []PlayStatusDiscrepancyDTO{
					{UserID: "u1", UserName: "User1", MovieToUpdate: "m2", MovieName: "Movie"},
				},
			},
			expectedDuplicate: false,
		},
		{
			name: "Duplicate with discrepancies",
			duplicate: DuplicateResultDTO{
				IsDuplicate:              true,
				Similarity:               95,
				HasIdenticalPlayStatus:   false,
				HasPlayStatusDiscrepancy: true,
				PlayStatusDiscrepancies: []PlayStatusDiscrepancyDTO{
					{UserID: "u1", UserName: "User1", MovieToUpdate: "m2", MovieName: "Movie"},
					{UserID: "u2", UserName: "User2", MovieToUpdate: "m1", MovieName: "Movie"},
				},
			},
			expectedDuplicate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedDuplicate, tt.duplicate.IsDuplicate)
			assert.Greater(t, tt.duplicate.Similarity, 0)
			assert.IsType(t, []PlayStatusDiscrepancyDTO{}, tt.duplicate.PlayStatusDiscrepancies)
		})
	}
}

// TestMovieLightStatusDTO tests the MovieLightStatusDTO struct
func TestMovieLightStatusDTO(t *testing.T) {
	tests := []struct {
		name  string
		movie MovieLightStatusDTO
	}{
		{
			name: "Movie with user statuses",
			movie: MovieLightStatusDTO{
				UserPlayStatuses: []UserPlayStatusDTO{
					{UserID: "u1", UserName: "User1", Played: true, PlayCount: 2},
					{UserID: "u2", UserName: "User2", Played: false, PlayCount: 0},
				},
			},
		},
		{
			name: "Movie with single user",
			movie: MovieLightStatusDTO{
				UserPlayStatuses: []UserPlayStatusDTO{
					{UserID: "u1", UserName: "User1", Played: true, PlayCount: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.movie.UserPlayStatuses)
			assert.Greater(t, len(tt.movie.UserPlayStatuses), 0)
		})
	}
}

// TestMovieDTO tests the MovieDTO struct
func TestMovieDTO(t *testing.T) {
	tests := []struct {
		name  string
		movie MovieDTO
	}{
		{
			name: "Full movie DTO",
			movie: MovieDTO{
				UserPlayStatuses: []UserPlayStatusDTO{
					{UserID: "u1", UserName: "User1", Played: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.movie.UserPlayStatuses)
		})
	}
}

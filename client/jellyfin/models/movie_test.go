package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMovieXtLightAPI tests the MovieXtLightAPI struct
func TestMovieXtLightAPI(t *testing.T) {
	tests := []struct {
		name         string
		movie        MovieXtLightAPI
		expectedID   string
		expectedName string
		expectedYear int
	}{
		{
			name: "Movie with all fields",
			movie: MovieXtLightAPI{
				ID:             "movie1",
				Name:           "Test Movie",
				ProductionYear: 2023,
			},
			expectedID:   "movie1",
			expectedName: "Test Movie",
			expectedYear: 2023,
		},
		{
			name: "Movie with zero year",
			movie: MovieXtLightAPI{
				ID:             "movie2",
				Name:           "Unknown Year Movie",
				ProductionYear: 0,
			},
			expectedID:   "movie2",
			expectedName: "Unknown Year Movie",
			expectedYear: 0,
		},
		{
			name: "Movie with empty fields",
			movie: MovieXtLightAPI{
				ID:             "",
				Name:           "",
				ProductionYear: 2020,
			},
			expectedID:   "",
			expectedName: "",
			expectedYear: 2020,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedID, tt.movie.ID)
			assert.Equal(t, tt.expectedName, tt.movie.Name)
			assert.Equal(t, tt.expectedYear, tt.movie.ProductionYear)
		})
	}
}

// TestMovieLightAPI tests the MovieLightAPI struct
func TestMovieLightAPI(t *testing.T) {
	tests := []struct {
		name  string
		movie MovieLightAPI
	}{
		{
			name: "Movie light with path",
			movie: MovieLightAPI{
				MovieXtLightAPI: MovieXtLightAPI{
					ID:             "light1",
					Name:           "Light Movie",
					ProductionYear: 2023,
				},
				Path: "/path/to/movie",
			},
		},
		{
			name: "Movie light without path",
			movie: MovieLightAPI{
				MovieXtLightAPI: MovieXtLightAPI{
					ID:             "light2",
					Name:           "No Path Movie",
					ProductionYear: 2022,
				},
				Path: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.movie.MovieXtLightAPI)
			assert.NotEmpty(t, tt.movie.MovieXtLightAPI.ID)
		})
	}
}

// TestUserDataAPI tests the UserDataAPI struct
func TestUserDataAPI(t *testing.T) {
	tests := []struct {
		name              string
		userData          UserDataAPI
		expectedPlayed    bool
		expectedPlayCount int
	}{
		{
			name: "Played movie with count",
			userData: UserDataAPI{
				Played:                true,
				PlaybackPositionTicks: 1000000,
				PlayCount:             3,
				LastPlayedDate:        "2023-01-15T10:00:00Z",
			},
			expectedPlayed:    true,
			expectedPlayCount: 3,
		},
		{
			name: "Not played movie",
			userData: UserDataAPI{
				Played:                false,
				PlaybackPositionTicks: 0,
				PlayCount:             0,
				LastPlayedDate:        "",
			},
			expectedPlayed:    false,
			expectedPlayCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedPlayed, tt.userData.Played)
			assert.Equal(t, tt.expectedPlayCount, tt.userData.PlayCount)
		})
	}
}

// TestMediaStreamsAPI tests the MediaStreamsAPI struct
func TestMediaStreamsAPI(t *testing.T) {
	tests := []struct {
		name         string
		stream       MediaStreamsAPI
		expectedType string
	}{
		{
			name: "Video stream",
			stream: MediaStreamsAPI{
				DisplayTitle: "H.264 1080p",
				Type:         "Video",
			},
			expectedType: "Video",
		},
		{
			name: "Audio stream",
			stream: MediaStreamsAPI{
				DisplayTitle: "AAC 2.0",
				Type:         "Audio",
			},
			expectedType: "Audio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedType, tt.stream.Type)
			assert.NotEmpty(t, tt.stream.DisplayTitle)
		})
	}
}

// TestMovieAPI tests the MovieAPI struct
func TestMovieAPI(t *testing.T) {
	tests := []struct {
		name  string
		movie MovieAPI
	}{
		{
			name: "Movie with user data and streams",
			movie: MovieAPI{
				MovieXtLightAPI: MovieXtLightAPI{
					ID:             "full1",
					Name:           "Full Movie",
					ProductionYear: 2023,
				},
				UserData: UserDataAPI{
					Played:    true,
					PlayCount: 2,
				},
				MediaStreams: []MediaStreamsAPI{
					{DisplayTitle: "Video", Type: "Video"},
					{DisplayTitle: "Audio", Type: "Audio"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.movie.UserData.Played)
			assert.Greater(t, len(tt.movie.MediaStreams), 0)
		})
	}
}

// TestMovieLightExtendedAPI tests the MovieLightExtendedAPI struct
func TestMovieLightExtendedAPI(t *testing.T) {
	tests := []struct {
		name  string
		movie MovieLightExtendedAPI
	}{
		{
			name: "Extended movie with streams",
			movie: MovieLightExtendedAPI{
				MovieLightAPI: MovieLightAPI{
					MovieXtLightAPI: MovieXtLightAPI{
						ID:             "ext1",
						Name:           "Extended",
						ProductionYear: 2023,
					},
					Path: "/extended/path",
				},
				MediaStreams: []MediaStreamsAPI{
					{DisplayTitle: "Video", Type: "Video"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.movie.MovieLightAPI.Path)
			assert.NotEmpty(t, tt.movie.MediaStreams)
		})
	}
}

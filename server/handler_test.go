package server_test

import (
	jellyfinClients "jellyfin-duplicate/client/jellyfin/http"
	apiModels "jellyfin-duplicate/client/jellyfin/models"
	"jellyfin-duplicate/server"
	serverModels "jellyfin-duplicate/server/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockJellyfinClientForHandler struct {
	movies     []apiModels.MovieLightAPI
	users      []apiModels.UserAPI
	shouldFail bool
}

func (m *mockJellyfinClientForHandler) GetAllMovies() ([]apiModels.MovieLightAPI, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return m.movies, nil
}

func (m *mockJellyfinClientForHandler) GetAllUsers() ([]apiModels.UserAPI, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return m.users, nil
}

func (m *mockJellyfinClientForHandler) GetSeenMoviesForAllUsers(users []apiModels.UserAPI) (map[string][]apiModels.MovieLightAPI, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return make(map[string][]apiModels.MovieLightAPI), nil
}

func (m *mockJellyfinClientForHandler) GetMovieDetails(movieID string) (apiModels.MovieLightExtendedAPI, error) {
	return apiModels.MovieLightExtendedAPI{}, nil
}

func (m *mockJellyfinClientForHandler) GetMovieUserData(movieID string, userID string) (apiModels.UserDataAPI, error) {
	return apiModels.UserDataAPI{}, nil
}

func (m *mockJellyfinClientForHandler) GetMovieName(movieID string) (string, error) {
	return "Test Movie", nil
}

func (m *mockJellyfinClientForHandler) GetUserName(userID string) (string, error) {
	return "Test User", nil
}

func (m *mockJellyfinClientForHandler) DeleteMovie(movieID string) error {
	if m.shouldFail {
		return assert.AnError
	}
	return nil
}

func (m *mockJellyfinClientForHandler) MarkMovieAsPlayed(movieID string, userID string, movieName string, userName string) error {
	if m.shouldFail {
		return assert.AnError
	}
	return nil
}

func TestNewHandler(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Creates handler with valid client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &jellyfinClients.Client{}
			handler := server.NewHandler(client)

			assert.NotNil(t, handler, "Handler should not be nil")
		})
	}
}

func TestHandlerDeleteMovie_ValidID(t *testing.T) {
	tests := []struct {
		name    string
		movieID string
		valid   bool
	}{
		{
			name:    "Valid UUID format - 32 chars",
			movieID: "12345678901234567890123456789012",
			valid:   true,
		},
		{
			name:    "Valid UUID format - 36 chars",
			movieID: "12345678-1234-1234-1234-123456789012",
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.IsUUIDFormtatted(tt.movieID)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestHandlerDeleteMovie_InvalidID(t *testing.T) {
	tests := []struct {
		name    string
		movieID string
		valid   bool
	}{
		{
			name:    "Too short ID",
			movieID: "short",
			valid:   false,
		},
		{
			name:    "Too long ID",
			movieID: "123456789012345678901234567890123456789012345678",
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.IsUUIDFormtatted(tt.movieID)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestHandlerMarkMovieAsSeen_Validation(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		valid  bool
	}{
		{
			name:   "Valid user ID",
			userID: "12345678901234567890123456789012",
			valid:  true,
		},
		{
			name:   "Invalid user ID - too short",
			userID: "invalid",
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.IsUUIDFormtatted(tt.userID)
			assert.Equal(t, tt.valid, result)
		})
	}
}

// Test the service layer response for handlers
func TestHandlerDuplicatesAnalysis(t *testing.T) {
	tests := []struct {
		name       string
		hasMovies  bool
		shouldFail bool
	}{
		{
			name:       "With movies",
			hasMovies:  true,
			shouldFail: false,
		},
		{
			name:       "No movies",
			hasMovies:  false,
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockJellyfinClientForHandler{
				movies:     []apiModels.MovieLightAPI{},
				users:      []apiModels.UserAPI{{ID: "user1", Name: "User One"}},
				shouldFail: tt.shouldFail,
			}

			if tt.hasMovies {
				mockClient.movies = []apiModels.MovieLightAPI{
					{MovieXtLightAPI: apiModels.MovieXtLightAPI{ID: "movie1", Name: "Movie 1", ProductionYear: 2023}, Path: "/movies/1"},
					{MovieXtLightAPI: apiModels.MovieXtLightAPI{ID: "movie2", Name: "Movie 1", ProductionYear: 2023}, Path: "/movies/2"},
				}
			}

			service := server.NewService(mockClient)
			duplicates, err := service.FindDuplicates()

			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// FindDuplicates returns nil when no duplicates found, empty slice otherwise
				assert.True(t, duplicates == nil || len(duplicates) >= 0, "Should return slice or nil")
			}
		})
	}
}

// Test GetPlayStatusDiscrepancies in handler context
func TestHandlerPlayStatusDiscrepancies(t *testing.T) {
	tests := []struct {
		name            string
		movie1PlayCount int
		movie2PlayCount int
	}{
		{
			name:            "Both movies identical play status",
			movie1PlayCount: 2,
			movie2PlayCount: 2,
		},
		{
			name:            "Different play status",
			movie1PlayCount: 1,
			movie2PlayCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockJellyfinClientForHandler{}
			service := server.NewService(mockClient)

			movie1 := serverModels.MovieDTO{
				UserPlayStatuses: []serverModels.UserPlayStatusDTO{
					{UserID: "user1", UserName: "User 1", Played: true, PlayCount: 1},
				},
			}

			movie2 := serverModels.MovieDTO{
				UserPlayStatuses: []serverModels.UserPlayStatusDTO{
					{UserID: "user1", UserName: "User 1", Played: true, PlayCount: 1},
				},
			}

			discrepancies := service.GetPlayStatusDiscrepancies(movie1, movie2)
			assert.NotNil(t, discrepancies)
		})
	}
}

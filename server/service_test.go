package server_test

import (
	"fmt"
	jellyfinClients "jellyfin-duplicate/client/jellyfin/http"
	apiModels "jellyfin-duplicate/client/jellyfin/models"
	"jellyfin-duplicate/server"
	serverModels "jellyfin-duplicate/server/models"
	"jellyfin-duplicate/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerService_GetPlayStatusDiscrepancies(t *testing.T) {

	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			movie1 := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/movie1.json", functionName, useCase), serverModels.MovieDTO{})
			movie2 := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/movie2.json", functionName, useCase), serverModels.MovieDTO{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), []serverModels.PlayStatusDiscrepancyDTO{})
			// Execution
			got := server.NewService(&jellyfinClients.Client{}).GetPlayStatusDiscrepancies(movie1, movie2)
			// Validation
			assert.Equal(t, expected, got, "Should return the expected play status discrepancies")
		})
	}
}

func TestIsUUIDFormtatted(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		id   string
		want bool
	}{
		{
			name: "Valid UUID",
			id:   "123e4567-e89b-12d3-a456-426614174000",
			want: true,
		},
		{
			name: "Valid UUID - missing hyphens",
			id:   "123e4567e89b12d3a456426614174000",
			want: true,
		},
		{
			name: "Invalid UUID - wrong length",
			id:   "123e4567-e89b-12d3-a456-42661417400-56456465",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := server.IsUUIDFormtatted(tt.id)
			assert.Equal(t, tt.want, got, "Should return whether the string is a valid UUID format")
		})
	}
}

func TestServerService_ReconcilePlayStatusWithAllMovies(t *testing.T) {
	tests := []struct {
		name           string // description of this test case
		client         *jellyfinClients.Client
		allMovies      []apiModels.MovieLightAPI
		userSeenMovies map[string][]apiModels.MovieLightAPI
		users          []apiModels.UserAPI
		want           []serverModels.MovieLightStatusDTO
		wantErr        bool
	}{
		{
			name:   "No movies",
			client: &jellyfinClients.Client{},
			want:   []serverModels.MovieLightStatusDTO(nil),
		},
		{
			name:   "One movie, one user, seen",
			client: &jellyfinClients.Client{},
			allMovies: []apiModels.MovieLightAPI{
				{
					MovieXtLightAPI: apiModels.MovieXtLightAPI{
						ID:   "1",
						Name: "movie1",
					},
				},
			},
			userSeenMovies: map[string][]apiModels.MovieLightAPI{
				"user1": {
					{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "1",
							Name: "movie1",
						},
					},
				},
			},
			users: []apiModels.UserAPI{
				{

					ID:   "user1",
					Name: "User One",
				},
			},
			want: []serverModels.MovieLightStatusDTO{
				{
					MovieLightAPI: apiModels.MovieLightAPI{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "1",
							Name: "movie1",
						},
					},
					UserPlayStatuses: []serverModels.UserPlayStatusDTO{
						{
							UserID:    "user1",
							UserName:  "User One",
							Played:    true,
							PlayCount: 0,
						},
					},
				},
			},
		},
		{
			name:   "One movie, one user, not seen",
			client: &jellyfinClients.Client{},
			allMovies: []apiModels.MovieLightAPI{
				{
					MovieXtLightAPI: apiModels.MovieXtLightAPI{
						ID:   "1",
						Name: "movie1",
					},
				},
			},
			userSeenMovies: map[string][]apiModels.MovieLightAPI{
				"user1": {},
			},
			users: []apiModels.UserAPI{
				{

					ID:   "user1",
					Name: "User One",
				},
			},
			want: []serverModels.MovieLightStatusDTO{
				{
					MovieLightAPI: apiModels.MovieLightAPI{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "1",
							Name: "movie1",
						},
					},
					UserPlayStatuses: []serverModels.UserPlayStatusDTO{
						{
							UserID:    "user1",
							UserName:  "User One",
							Played:    false,
							PlayCount: 0,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "Multiple movies, multiple users",
			client: &jellyfinClients.Client{},
			allMovies: []apiModels.MovieLightAPI{
				{
					MovieXtLightAPI: apiModels.MovieXtLightAPI{
						ID:   "1",
						Name: "movie1",
					},
				},
				{
					MovieXtLightAPI: apiModels.MovieXtLightAPI{
						ID:   "2",
						Name: "movie2",
					},
				},
			},
			userSeenMovies: map[string][]apiModels.MovieLightAPI{
				"user1": {
					{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "1",
							Name: "movie1",
						},
					},
				},
				"user2": {
					{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "2",
							Name: "movie2",
						},
					},
				},
			},
			users: []apiModels.UserAPI{
				{

					ID:   "user1",
					Name: "User One",
				},
				{

					ID:   "user2",
					Name: "User Two",
				},
			},
			want: []serverModels.MovieLightStatusDTO{
				{
					MovieLightAPI: apiModels.MovieLightAPI{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "1",
							Name: "movie1",
						},
					},
					UserPlayStatuses: []serverModels.UserPlayStatusDTO{
						{
							UserID:    "user1",
							UserName:  "User One",
							Played:    true,
							PlayCount: 0,
						},
						{
							UserID:    "user2",
							UserName:  "User Two",
							Played:    false,
							PlayCount: 0,
						},
					},
				},
				{
					MovieLightAPI: apiModels.MovieLightAPI{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "2",
							Name: "movie2",
						},
					},
					UserPlayStatuses: []serverModels.UserPlayStatusDTO{
						{
							UserID:    "user1",
							UserName:  "User One",
							Played:    false,
							PlayCount: 0,
						},
						{
							UserID:    "user2",
							UserName:  "User Two",
							Played:    true,
							PlayCount: 0,
						},
					},
				},
			},
		},
		{
			name:   "Movie seen by multiple users",
			client: &jellyfinClients.Client{},
			allMovies: []apiModels.MovieLightAPI{
				{
					MovieXtLightAPI: apiModels.MovieXtLightAPI{
						ID:   "1",
						Name: "movie1",
					},
				},
			},
			userSeenMovies: map[string][]apiModels.MovieLightAPI{
				"user1": {
					{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "1",
							Name: "movie1",
						},
					},
				},
				"user2": {
					{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "1",
							Name: "movie1",
						},
					},
				},
			},
			users: []apiModels.UserAPI{
				{

					ID:   "user1",
					Name: "User One",
				},
				{

					ID:   "user2",
					Name: "User Two",
				},
			},
			want: []serverModels.MovieLightStatusDTO{
				{
					MovieLightAPI: apiModels.MovieLightAPI{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "1",
							Name: "movie1",
						},
					},
					UserPlayStatuses: []serverModels.UserPlayStatusDTO{
						{
							UserID:    "user1",
							UserName:  "User One",
							Played:    true,
							PlayCount: 0,
						},
						{
							UserID:    "user2",
							UserName:  "User Two",
							Played:    true,
							PlayCount: 0,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := server.NewService(tt.client)
			got, gotErr := s.ReconcilePlayStatusWithAllMovies(tt.allMovies, tt.userSeenMovies, tt.users)
			if gotErr != nil {
				assert.Nil(t, tt.wantErr, "Expected no error but got: %v", gotErr)
			}
			assert.NotNil(t, tt.wantErr, "Expected an error but got nil")
			assert.Equal(t, tt.want, got)
		})
	}
}

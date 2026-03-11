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
			assert.ElementsMatch(t, expected, got, "Should return the expected play status discrepancies")
		})
	}
}

func TestServerService_ReconcilePlayStatusWithAllMovies(t *testing.T) {

	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			allMovies := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/all_movies.json", functionName, useCase), []apiModels.MovieLightAPI{})
			userSeenMovies := make(map[string][]apiModels.MovieLightAPI)
			userSeenMovies["user1"] = test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/user1_seen_movies.json", functionName, useCase), []apiModels.MovieLightAPI{})
			userSeenMovies["user2"] = test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/user2_seen_movies.json", functionName, useCase), []apiModels.MovieLightAPI{})
			users := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/users.json", functionName, useCase), []apiModels.UserAPI{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), []serverModels.MovieLightStatusDTO{})
			// Execution
			got, gotErr := server.NewService(&jellyfinClients.Client{}).ReconcilePlayStatusWithAllMovies(allMovies, userSeenMovies, users)
			// Validation
			assert.ElementsMatch(t, expected, got, "Should return the expected reconciled play status for all movies and users")
			assert.Nil(t, gotErr, "Expected no error but got: %v", gotErr)
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

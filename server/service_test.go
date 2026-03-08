package server_test

import (
	jellyfinClients "jellyfin-duplicate/client/jellyfin/http"
	apiModels "jellyfin-duplicate/client/jellyfin/models"
	"jellyfin-duplicate/server"
	serverModels "jellyfin-duplicate/server/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerService_GetPlayStatusDiscrepancies(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		client *jellyfinClients.Client
		// Named input parameters for target function.
		movie1 serverModels.MovieDTO
		movie2 serverModels.MovieDTO
		want   []serverModels.PlayStatusDiscrepancyDTO
	}{
		{
			name:   "No discrepancies",
			client: &jellyfinClients.Client{},
			movie1: serverModels.MovieDTO{
				UserPlayStatuses: []serverModels.UserPlayStatusDTO{
					{
						UserID:    "user1",
						UserName:  "Played",
						Played:    true,
						PlayCount: 1,
					},
				},
			},
			movie2: serverModels.MovieDTO{
				UserPlayStatuses: []serverModels.UserPlayStatusDTO{
					{
						UserID:    "user1",
						UserName:  "Played",
						Played:    true,
						PlayCount: 1,
					},
				},
			},
			want: []serverModels.PlayStatusDiscrepancyDTO(nil),
		},
		{
			name:   "One discrepancy",
			client: &jellyfinClients.Client{},
			movie1: serverModels.MovieDTO{
				MovieLightExtendedAPI: apiModels.MovieLightExtendedAPI{
					MovieLightAPI: apiModels.MovieLightAPI{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "1",
							Name: "movie1",
						},
					},
				},
				UserPlayStatuses: []serverModels.UserPlayStatusDTO{
					{
						UserID:    "user1",
						UserName:  "Played",
						Played:    true,
						PlayCount: 1,
					},
				},
			},
			movie2: serverModels.MovieDTO{
				MovieLightExtendedAPI: apiModels.MovieLightExtendedAPI{
					MovieLightAPI: apiModels.MovieLightAPI{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "2",
							Name: "movie2",
						},
					},
				},
				UserPlayStatuses: []serverModels.UserPlayStatusDTO{
					{
						UserID:    "user1",
						UserName:  "Played",
						Played:    false,
						PlayCount: 0,
					},
				},
			},
			want: []serverModels.PlayStatusDiscrepancyDTO{
				{
					UserID:        "user1",
					UserName:      "Played",
					MovieToUpdate: "2",
					MovieName:     "movie2",
				},
			},
		},
		{
			name:   "One discrepancy second movie",
			client: &jellyfinClients.Client{},
			movie1: serverModels.MovieDTO{
				MovieLightExtendedAPI: apiModels.MovieLightExtendedAPI{
					MovieLightAPI: apiModels.MovieLightAPI{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "1",
							Name: "movie1",
						},
					},
				},
				UserPlayStatuses: []serverModels.UserPlayStatusDTO{
					{
						UserID:    "user1",
						UserName:  "Played",
						Played:    true,
						PlayCount: 1,
					},
				},
			},
			movie2: serverModels.MovieDTO{
				MovieLightExtendedAPI: apiModels.MovieLightExtendedAPI{
					MovieLightAPI: apiModels.MovieLightAPI{
						MovieXtLightAPI: apiModels.MovieXtLightAPI{
							ID:   "2",
							Name: "movie2",
						},
					},
				},
				UserPlayStatuses: []serverModels.UserPlayStatusDTO{
					{
						UserID:    "user2",
						UserName:  "Played",
						Played:    true,
						PlayCount: 1,
					},
				},
			},
			want: []serverModels.PlayStatusDiscrepancyDTO{
				{
					UserID:        "user1",
					UserName:      "Played",
					MovieToUpdate: "2",
					MovieName:     "movie2",
				},
				{
					UserID:        "user2",
					UserName:      "Played",
					MovieToUpdate: "1",
					MovieName:     "movie1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := server.NewService(tt.client)
			got := s.GetPlayStatusDiscrepancies(tt.movie1, tt.movie2)
			assert.Equal(t, tt.want, got, "Should return the expected play status discrepancies")
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

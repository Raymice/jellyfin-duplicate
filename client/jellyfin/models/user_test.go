package models_test

import (
	"jellyfin-duplicate/client/jellyfin/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUserAPI tests the UserAPI struct
func TestUserAPI(t *testing.T) {
	tests := []struct {
		name                string
		user                models.UserAPI
		expectedHasPassword bool
	}{
		{
			name: "User with password",
			user: models.UserAPI{
				ID:               "user1",
				Name:             "John Doe",
				HasPassword:      true,
				LastLoginDate:    "2023-01-15T10:00:00Z",
				LastActivityDate: "2023-01-15T11:00:00Z",
			},
			expectedHasPassword: true,
		},
		{
			name: "User without password",
			user: models.UserAPI{
				ID:               "user2",
				Name:             "Jane Doe",
				HasPassword:      false,
				LastLoginDate:    "",
				LastActivityDate: "",
			},
			expectedHasPassword: false,
		},
		{
			name: "User with empty name",
			user: models.UserAPI{
				ID:          "user3",
				Name:        "",
				HasPassword: true,
			},
			expectedHasPassword: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedHasPassword, tt.user.HasPassword)
			assert.NotEmpty(t, tt.user.ID)
		})
	}
}

// TestLibraryAPI tests the LibraryAPI struct
func TestLibraryAPI(t *testing.T) {
	tests := []struct {
		name         string
		library      models.LibraryAPI
		expectedName string
	}{
		{
			name: "Movies library",
			library: models.LibraryAPI{
				ID:   "lib1",
				Name: "Movies",
			},
			expectedName: "Movies",
		},
		{
			name: "TV Shows library",
			library: models.LibraryAPI{
				ID:   "lib2",
				Name: "TV Shows",
			},
			expectedName: "TV Shows",
		},
		{
			name: "Library with special characters",
			library: models.LibraryAPI{
				ID:   "lib3",
				Name: "My Films & Videos",
			},
			expectedName: "My Films & Videos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedName, tt.library.Name)
			assert.NotEmpty(t, tt.library.ID)
		})
	}
}

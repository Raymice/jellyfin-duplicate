//go:build e2e

package http

import (
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"jellyfin-duplicate/constants"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

// TestJellyfinE2E exercises read-only API calls against the server in the root .env.
func TestJellyfinE2E(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	require.True(t, ok, "locate test source")
	envPath := filepath.Join(filepath.Dir(source), "..", "..", "..", ".env")
	settings, err := godotenv.Read(envPath)
	require.NoError(t, err, "read repository .env")
	for _, key := range []string{constants.EnvJellyfinURL, constants.EnvJellyfinAPIKey, constants.EnvJellyfinAdminUserID} {
		require.NotEmpty(t, settings[key], "missing %s in .env", key)
	}
	baseURL := strings.TrimRight(settings[constants.EnvJellyfinURL], "/")
	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err, "parse JELLYFIN_URL")
	require.Contains(t, []string{"http", "https"}, parsedURL.Scheme, "JELLYFIN_URL must use HTTP or HTTPS")
	require.NotEmpty(t, parsedURL.Host, "JELLYFIN_URL must include a host")

	userID := settings[constants.EnvJellyfinAdminUserID]
	client := NewClient(baseURL, settings[constants.EnvJellyfinAPIKey], userID)
	client.client.SetTimeout(30 * time.Second)
	require.NoError(t, client.ConfigureAuthentication(settings[constants.EnvJellyfinAuth]))

	t.Run("public server version", func(t *testing.T) {
		version, err := client.GetServerVersion()
		require.NoError(t, err)
		require.NotEmpty(t, version)
		t.Logf("Jellyfin server version: %s", version)
	})

	t.Run("users", func(t *testing.T) {
		users, err := client.GetAllUsers()
		require.NoError(t, err)
		// Jellyfin may return UUIDs without hyphens.
		for _, user := range users {
			if strings.EqualFold(strings.ReplaceAll(user.ID, "-", ""), strings.ReplaceAll(userID, "-", "")) {
				return
			}
		}
		t.Fatal("configured admin user was not returned by Jellyfin")
	})

	libraries, err := client.GetLibraries()
	require.NoError(t, err, "fetch libraries for configured user")
	t.Run("seen movies", func(t *testing.T) {
		_, err := client.GetSeenMoviesForUser(userID)
		require.NoError(t, err)
	})

	t.Run("movie reads", func(t *testing.T) {
		for _, library := range libraries {
			movies, err := client.getMoviesFromLibrary(library.ID)
			require.NoError(t, err, "fetch library movies")
			if len(movies) == 0 {
				continue
			}
			movieID := movies[0].ID
			require.NotEmpty(t, movieID)
			details, err := client.GetMovieDetails(movieID)
			require.NoError(t, err, "fetch movie details")
			require.Equal(t, movieID, details.ID)
			name, err := client.GetMovieName(movieID)
			require.NoError(t, err, "fetch movie name")
			require.NotEmpty(t, name)
			_, err = client.GetMovieUserData(movieID, userID)
			require.NoError(t, err, "fetch movie watch history")
			return
		}
		t.Skip("no movies available; authentication and library checks still ran")
	})
}

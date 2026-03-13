package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func mockServer(response string, statusCode int) *httptest.Server {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(response))
	}))

	return server
}

func mockClient() *Client {
	return &Client{
		baseURL:   "http://localhost:8096",
		apiKey:    "test-api-key",
		userID:    "current-user",
		client:    resty.New(),
		userCache: make(map[string]string),
	}
}

func TestGetUserName(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		cachedName    string
		apiResponse   string
		expectedName  string
		expectError   bool
		errorContains string
	}{
		{
			name:         "Returns cached user name",
			userID:       "user-123",
			cachedName:   "John Doe",
			expectedName: "John Doe",
			expectError:  false,
		},
		{
			name:         "Fetches and caches user name from API",
			userID:       "user-456",
			cachedName:   "",
			apiResponse:  `{"Name":"Jane Smith"}`,
			expectedName: "Jane Smith",
			expectError:  false,
		},
		{
			name:          "Handles API error",
			userID:        "user-789",
			cachedName:    "",
			apiResponse:   ``,
			expectedName:  "",
			expectError:   true,
			errorContains: "failed to fetch user name",
		},
		{
			name:         "Handles empty API response",
			userID:       "user-empty",
			cachedName:   "",
			apiResponse:  `{"Name":""}`,
			expectedName: "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mockClient()

			// Pre-populate cache if needed
			if tt.cachedName != "" {
				client.userCache[tt.userID] = tt.cachedName
			}

			// Mock the HTTP client if API call is expected
			var server *httptest.Server
			if tt.expectError {
				server = mockServer(`{"Message":"Internal Server Error"}`, http.StatusInternalServerError)
			} else {
				server = mockServer(tt.apiResponse, http.StatusOK)
			}

			defer server.Close()
			client.baseURL = server.URL

			result, err := client.GetUserName(tt.userID)

			if tt.expectError {
				assert.Error(t, err, "GetUserName should return an error")
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err, "GetUserName should not return an error")
				assert.Equal(t, tt.expectedName, result, "GetUserName returned unexpected result")
			}
		})
	}
}

func TestDeleteMovie(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		response      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "Delete 204 No Content",
			statusCode:  http.StatusNoContent,
			response:    "",
			expectError: false,
		},
		{
			name:        "Delete 200 OK",
			statusCode:  http.StatusOK,
			response:    "",
			expectError: false,
		},
		{
			name:          "Delete not found returns error",
			statusCode:    http.StatusNotFound,
			response:      `{"Message":"Not Found"}`,
			expectError:   true,
			errorContains: "unexpected status code 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mockClient()
			server := mockServer(tt.response, tt.statusCode)
			defer server.Close()

			client.baseURL = server.URL

			err := client.DeleteMovie("movie-123")

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestNewClient(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		apiKey string
		userID string
	}{
		{
			name:   "Creates client with valid parameters",
			url:    "http://localhost:8096",
			apiKey: "test-api-key",
			userID: "user-123",
		},
		{
			name:   "Creates client with empty string values",
			url:    "",
			apiKey: "",
			userID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.url, tt.apiKey, tt.userID)

			assert.NotNil(t, client)
			assert.Equal(t, tt.url, client.baseURL)
			assert.Equal(t, tt.apiKey, client.apiKey)
			assert.Equal(t, tt.userID, client.userID)
			assert.NotNil(t, client.client)
			assert.NotNil(t, client.userCache)
			assert.Equal(t, 0, len(client.userCache))
		})
	}
}

func TestGetLibraries(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		apiResponse   string
		statusCode    int
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:          "Fetches libraries successfully",
			userID:        "user-123",
			apiResponse:   `{"Items": [{"Id": "lib-1", "Name": "Movies"}, {"Id": "lib-2", "Name": "TV Shows"}]}`,
			statusCode:    http.StatusOK,
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:          "Handles empty libraries",
			userID:        "user-123",
			apiResponse:   `{"Items": []}`,
			statusCode:    http.StatusOK,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:          "Handles API error",
			userID:        "user-123",
			apiResponse:   `{"Message": "Internal Server Error"}`,
			statusCode:    http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to fetch libraries",
		},
		{
			name:          "Handles missing user ID",
			userID:        "",
			apiResponse:   ``,
			statusCode:    http.StatusOK,
			expectError:   true,
			errorContains: "user ID not set",
		},
		{
			name:          "Handles invalid JSON response",
			userID:        "user-123",
			apiResponse:   `invalid json`,
			statusCode:    http.StatusOK,
			expectError:   true,
			errorContains: "failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mockClient()
			if tt.userID != "" {
				client.userID = tt.userID
			} else {
				client.userID = ""
			}

			if tt.userID != "" {
				server := mockServer(tt.apiResponse, tt.statusCode)
				defer server.Close()
				client.baseURL = server.URL
			}

			result, err := client.GetLibraries()

			if tt.expectError {
				assert.Error(t, err, "GetLibraries should return an error")
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(result))
			}
		})
	}
}

func TestGetAllUsers(t *testing.T) {
	tests := []struct {
		name          string
		apiResponse   string
		statusCode    int
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:          "Fetches all users successfully",
			apiResponse:   `[{"Id": "user-1", "Name": "John Doe"}, {"Id": "user-2", "Name": "Jane Smith"}]`,
			statusCode:    http.StatusOK,
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:          "Fetches single user",
			apiResponse:   `[{"Id": "user-1", "Name": "John Doe"}]`,
			statusCode:    http.StatusOK,
			expectError:   false,
			expectedCount: 1,
		},
		{
			name:          "Handles empty user list",
			apiResponse:   `[]`,
			statusCode:    http.StatusOK,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:          "Handles API error",
			apiResponse:   `{"Message": "Internal Server Error"}`,
			statusCode:    http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to fetch users",
		},
		{
			name:        "Handles invalid JSON",
			apiResponse: `invalid json`,
			statusCode:  http.StatusOK,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mockClient()
			server := mockServer(tt.apiResponse, tt.statusCode)
			defer server.Close()
			client.baseURL = server.URL

			result, err := client.GetAllUsers()

			if tt.expectError {
				assert.Error(t, err, "GetAllUsers should return an error")
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(result))
				// Verify cache is populated
				assert.Equal(t, tt.expectedCount, len(client.userCache))
			}
		})
	}
}

func TestGetMovieUserData(t *testing.T) {
	tests := []struct {
		name          string
		movieID       string
		userID        string
		apiResponse   string
		statusCode    int
		expectError   bool
		errorContains string
		expectPlayed  bool
	}{
		{
			name:         "Fetches user data successfully (movie played)",
			movieID:      "movie-123",
			userID:       "user-456",
			apiResponse:  `{"Played": true, "PlayCount": 2, "LastPlayedDate": "2024-01-01"}`,
			statusCode:   http.StatusOK,
			expectError:  false,
			expectPlayed: true,
		},
		{
			name:         "Fetches user data successfully (movie not played)",
			movieID:      "movie-123",
			userID:       "user-456",
			apiResponse:  `{"Played": false, "PlayCount": 0}`,
			statusCode:   http.StatusOK,
			expectError:  false,
			expectPlayed: false,
		},
		{
			name:          "Handles API error",
			movieID:       "movie-123",
			userID:        "user-456",
			apiResponse:   `{"Message": "Not Found"}`,
			statusCode:    http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to fetch user play status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mockClient()
			server := mockServer(tt.apiResponse, tt.statusCode)
			defer server.Close()
			client.baseURL = server.URL

			result, err := client.GetMovieUserData(tt.movieID, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectPlayed, result.Played)
			}
		})
	}
}

func TestGetMovieDetails(t *testing.T) {
	tests := []struct {
		name          string
		movieID       string
		apiResponse   string
		statusCode    int
		expectError   bool
		errorContains string
		expectedName  string
	}{
		{
			name:         "Fetches movie details successfully",
			movieID:      "movie-123",
			apiResponse:  `{"Id": "movie-123", "Name": "The Matrix", "ProductionYear": 1999, "Path": "/movies/the-matrix.mkv"}`,
			statusCode:   http.StatusOK,
			expectError:  false,
			expectedName: "The Matrix",
		},
		{
			name:          "Handles movie not found",
			movieID:       "movie-999",
			apiResponse:   `{"Message": "Not Found"}`,
			statusCode:    http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to fetch movie details",
		},
		{
			name:          "Handles API error",
			movieID:       "movie-123",
			apiResponse:   `{"Message": "Internal Server Error"}`,
			statusCode:    http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to fetch movie details",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mockClient()
			server := mockServer(tt.apiResponse, tt.statusCode)
			defer server.Close()
			client.baseURL = server.URL

			result, err := client.GetMovieDetails(tt.movieID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, result.Name)
			}
		})
	}
}

func TestGetMovieName(t *testing.T) {
	tests := []struct {
		name          string
		movieID       string
		apiResponse   string
		statusCode    int
		expectError   bool
		errorContains string
		expectedName  string
	}{
		{
			name:         "Returns movie name successfully",
			movieID:      "movie-123",
			apiResponse:  `{"Name": "The Matrix"}`,
			statusCode:   http.StatusOK,
			expectError:  false,
			expectedName: "The Matrix",
		},
		{
			name:          "Handles movie not found",
			movieID:       "movie-999",
			apiResponse:   `{"Message": "Not Found"}`,
			statusCode:    http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to fetch movie name",
		},
		{
			name:         "Handles empty name and falls back",
			movieID:      "movie-123",
			apiResponse:  `{"Name": ""}`,
			statusCode:   http.StatusOK,
			expectError:  false,
			expectedName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mockClient()
			server := mockServer(tt.apiResponse, tt.statusCode)
			defer server.Close()
			client.baseURL = server.URL

			result, err := client.GetMovieName(tt.movieID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, result)
			}
		})
	}
}

func TestMarkMovieAsPlayed(t *testing.T) {
	tests := []struct {
		name          string
		movieID       string
		userID        string
		statusCode    int
		expectError   bool
		errorContains string
	}{
		{
			name:        "Marks movie as played with 204 response",
			movieID:     "movie-123",
			userID:      "user-456",
			statusCode:  http.StatusNoContent,
			expectError: false,
		},
		{
			name:        "Marks movie as played with 200 response",
			movieID:     "movie-123",
			userID:      "user-456",
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:          "Handles API error",
			movieID:       "movie-123",
			userID:        "user-456",
			statusCode:    http.StatusBadRequest,
			expectError:   true,
			errorContains: "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mockClient()
			server := mockServer("", tt.statusCode)
			defer server.Close()
			client.baseURL = server.URL

			err := client.MarkMovieAsPlayed(tt.movieID, tt.userID, "Test Movie", "Test User")

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckHTTPResponse(t *testing.T) {
	tests := []struct {
		name             string
		statusCode       int
		expectedStatuses []int
		expectError      bool
		errorContains    string
	}{
		{
			name:             "Success with matching status code",
			statusCode:       200,
			expectedStatuses: []int{200},
			expectError:      false,
		},
		{
			name:             "Success with multiple valid statuses",
			statusCode:       204,
			expectedStatuses: []int{200, 204},
			expectError:      false,
		},
		{
			name:             "Error with mismatched status code",
			statusCode:       400,
			expectedStatuses: []int{200},
			expectError:      true,
			errorContains:    "HTTP request failed",
		},
		{
			name:             "Error with unknown status code",
			statusCode:       418,
			expectedStatuses: []int{200},
			expectError:      true,
			errorContains:    "HTTP request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := resty.New()
			resp, err := client.R().Get(server.URL)
			assert.NoError(t, err)

			err = checkHTTPResponse(resp, tt.expectedStatuses...)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

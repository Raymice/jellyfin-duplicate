package http

import (
	"fmt"
	"jellyfin-duplicate/client/jellyfin/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"jellyfin-duplicate/test"

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

func TestNewClient(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				URL    string `json:"url"`
				APIKey string `json:"apiKey"`
				UserID string `json:"userId"`
			}{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), struct {
				BaseURL        string `json:"baseURL"`
				APIKey         string `json:"apiKey"`
				UserID         string `json:"userID"`
				UserCacheEmpty bool   `json:"userCacheEmpty"`
				HasClient      bool   `json:"hasClient"`
			}{})

			// Execution
			client := NewClient(input.URL, input.APIKey, input.UserID)

			// Validation
			assert.NotNil(t, client)
			assert.Equal(t, expected.BaseURL, client.baseURL)
			assert.Equal(t, expected.APIKey, client.apiKey)
			assert.Equal(t, expected.UserID, client.userID)
			if expected.HasClient {
				assert.NotNil(t, client.client)
			}
			if expected.UserCacheEmpty {
				assert.NotNil(t, client.userCache)
				assert.Equal(t, 0, len(client.userCache))
			}
		})
	}
}

func TestGetUserName(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				UserID      string `json:"userID"`
				CachedName  string `json:"cachedName"`
				APIResponse string `json:"apiResponse"`
				StatusCode  int    `json:"statusCode"`
				ExpectError bool   `json:"expectError"`
			}{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), struct {
				ExpectedName string `json:"expectedName"`
			}{})

			// Setup
			client := mockClient()
			if input.CachedName != "" {
				client.userCache[input.UserID] = input.CachedName
			}

			if input.APIResponse != "" {
				server := mockServer(input.APIResponse, input.StatusCode)
				defer server.Close()
				client.baseURL = server.URL
			}

			// Execution
			result, err := client.GetUserName(input.UserID)

			// Validation
			if input.ExpectError {
				assert.Error(t, err, "GetUserName should return an error")
			} else {
				assert.NoError(t, err, "GetUserName should not return an error")
				assert.Equal(t, expected.ExpectedName, result, "GetUserName returned unexpected result")
			}
		})
	}
}

func TestDeleteMovie(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				StatusCode  int    `json:"statusCode"`
				Response    string `json:"response"`
				ExpectError bool   `json:"expectError"`
			}{})

			// Setup
			client := mockClient()
			server := mockServer(input.Response, input.StatusCode)
			defer server.Close()
			client.baseURL = server.URL

			// Execution
			err := client.DeleteMovie("movie-123")

			// Validation
			if input.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetLibraries(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				UserID      string `json:"userID"`
				APIResponse string `json:"apiResponse"`
				StatusCode  int    `json:"statusCode"`
				ExpectError bool   `json:"expectError"`
			}{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), struct {
				ExpectedCount int `json:"expectedCount"`
			}{})

			// Setup
			client := mockClient()
			if input.UserID != "" {
				client.userID = input.UserID
			}

			if input.UserID != "" && input.APIResponse != "" {
				server := mockServer(input.APIResponse, input.StatusCode)
				defer server.Close()
				client.baseURL = server.URL
			}

			// Execution
			result, err := client.GetLibraries()

			// Validation
			if input.ExpectError {
				assert.Error(t, err, "GetLibraries should return an error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expected.ExpectedCount, len(result))
			}
		})
	}
}

func TestGetAllUsers(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				APIResponse string `json:"apiResponse"`
				StatusCode  int    `json:"statusCode"`
				ExpectError bool   `json:"expectError"`
			}{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), struct {
				ExpectedCount int `json:"expectedCount"`
			}{})

			// Setup
			client := mockClient()
			server := mockServer(input.APIResponse, input.StatusCode)
			defer server.Close()
			client.baseURL = server.URL

			// Execution
			result, err := client.GetAllUsers()

			// Validation
			if input.ExpectError {
				assert.Error(t, err, "GetAllUsers should return an error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expected.ExpectedCount, len(result))
				// Verify cache is populated
				assert.Equal(t, expected.ExpectedCount, len(client.userCache))
			}
		})
	}
}

func TestGetMovieUserData(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				MovieID     string `json:"movieID"`
				UserID      string `json:"userID"`
				APIResponse string `json:"apiResponse"`
				StatusCode  int    `json:"statusCode"`
				ExpectError bool   `json:"expectError"`
			}{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), struct {
				ExpectPlayed bool `json:"expectPlayed"`
			}{})

			// Setup
			client := mockClient()
			server := mockServer(input.APIResponse, input.StatusCode)
			defer server.Close()
			client.baseURL = server.URL

			// Execution
			result, err := client.GetMovieUserData(input.MovieID, input.UserID)

			// Validation
			if input.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expected.ExpectPlayed, result.Played)
			}
		})
	}
}

func TestGetMovieDetails(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				MovieID     string `json:"movieID"`
				APIResponse string `json:"apiResponse"`
				StatusCode  int    `json:"statusCode"`
				ExpectError bool   `json:"expectError"`
			}{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), struct {
				ExpectedName string `json:"expectedName"`
			}{})

			// Setup
			client := mockClient()
			server := mockServer(input.APIResponse, input.StatusCode)
			defer server.Close()
			client.baseURL = server.URL

			// Execution
			result, err := client.GetMovieDetails(input.MovieID)

			// Validation
			if input.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expected.ExpectedName, result.Name)
			}
		})
	}
}

func TestGetMovieName(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				MovieID     string `json:"movieID"`
				APIResponse string `json:"apiResponse"`
				StatusCode  int    `json:"statusCode"`
				ExpectError bool   `json:"expectError"`
			}{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), struct {
				ExpectedName string `json:"expectedName"`
			}{})

			// Setup
			client := mockClient()
			server := mockServer(input.APIResponse, input.StatusCode)
			defer server.Close()
			client.baseURL = server.URL

			// Execution
			result, err := client.GetMovieName(input.MovieID)

			// Validation
			if input.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expected.ExpectedName, result)
			}
		})
	}
}

func TestMarkMovieAsPlayed(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				MovieID     string `json:"movieID"`
				UserID      string `json:"userID"`
				StatusCode  int    `json:"statusCode"`
				ExpectError bool   `json:"expectError"`
			}{})

			// Setup
			client := mockClient()
			server := mockServer("", input.StatusCode)
			defer server.Close()
			client.baseURL = server.URL

			// Execution
			err := client.MarkMovieAsPlayed(input.MovieID, input.UserID, "Test Movie", "Test User")

			// Validation
			if input.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetAllMovies(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				LibrariesResponse string `json:"librariesResponse"`
				MoviesResponse    string `json:"moviesResponse"`
				LibrariesStatus   int    `json:"librariesStatus"`
				MoviesStatus      int    `json:"moviesStatus"`
				ExpectError       bool   `json:"expectError"`
				ExpectEmpty       bool   `json:"expectEmpty"`
			}{})

			// Setup
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/Users/user123/Views" {
					w.WriteHeader(input.LibrariesStatus)
					_, _ = w.Write([]byte(input.LibrariesResponse))
				} else {
					w.WriteHeader(input.MoviesStatus)
					_, _ = w.Write([]byte(input.MoviesResponse))
					requestCount++
				}
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", "user123")

			// Execution
			movies, err := client.GetAllMovies()

			// Validation
			if input.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if input.ExpectEmpty {
					assert.True(t, movies == nil || len(movies) == 0, "Should return empty or nil")
				} else {
					assert.Greater(t, len(movies), 0, "Should return at least one movie")
				}
			}
		})
	}
}

func TestGetSeenMoviesForUser(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				UserID      string `json:"userID"`
				Response    string `json:"response"`
				StatusCode  int    `json:"statusCode"`
				ExpectError bool   `json:"expectError"`
				ExpectEmpty bool   `json:"expectEmpty"`
			}{})

			// Setup
			server := mockServer(input.Response, input.StatusCode)
			defer server.Close()

			client := NewClient(server.URL, "test-key", "admin-user")

			// Execution
			movies, err := client.GetSeenMoviesForUser(input.UserID)

			// Validation
			if input.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if input.ExpectEmpty {
					// Empty response returns nil slice in Go
					assert.True(t, movies == nil || len(movies) == 0, "Movies should be nil or empty")
				} else {
					assert.NotNil(t, movies)
					assert.Greater(t, len(movies), 0)
				}
			}
		})
	}
}

func TestCheckHTTPResponse(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				StatusCode       int   `json:"statusCode"`
				ExpectedStatuses []int `json:"expectedStatuses"`
				ExpectError      bool  `json:"expectError"`
			}{})

			// Setup
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(input.StatusCode)
			}))
			defer server.Close()

			client := resty.New()
			resp, err := client.R().Get(server.URL)
			assert.NoError(t, err)

			// Execution
			err = checkHTTPResponse(resp, input.ExpectedStatuses...)

			// Validation
			if input.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSeenMoviesForAllUsers(t *testing.T) {
	tests := []struct {
		name         string
		users        []models.UserAPI
		mockResponse string
		statusCode   int
		expectError  bool
		expectMapLen int
		description  string
	}{
		{
			name: "Successfully fetches seen movies for multiple users",
			users: []models.UserAPI{
				{ID: "user1", Name: "User One"},
				{ID: "user2", Name: "User Two"},
				{ID: "user3", Name: "User Three"},
			},
			mockResponse: `{"Items":[{"Id":"movie1","Name":"Movie 1","ProductionYear":2020,"Path":"/movies/1"}],"TotalRecordCount":1}`,
			statusCode:   200,
			expectError:  false,
			expectMapLen: 3,
			description:  "All three users successfully return movies",
		},
		{
			name:         "Handles empty user list",
			users:        []models.UserAPI{},
			mockResponse: `{"Items":[],"TotalRecordCount":0}`,
			statusCode:   200,
			expectError:  false,
			expectMapLen: 0,
			description:  "Empty user list returns empty map",
		},
		{
			name: "Handles API error response",
			users: []models.UserAPI{
				{ID: "user1", Name: "User One"},
			},
			mockResponse: `{"error":"Internal Server Error"}`,
			statusCode:   500,
			expectError:  true,
			expectMapLen: 0,
			description:  "API error causes function to return error",
		},
		{
			name: "Successfully fetches for many concurrent users",
			users: []models.UserAPI{
				{ID: "user1", Name: "User One"},
				{ID: "user2", Name: "User Two"},
				{ID: "user3", Name: "User Three"},
				{ID: "user4", Name: "User Four"},
				{ID: "user5", Name: "User Five"},
				{ID: "user6", Name: "User Six"},
				{ID: "user7", Name: "User Seven"},
			},
			mockResponse: `{"Items":[{"Id":"movie1","Name":"Movie 1","ProductionYear":2020,"Path":"/movies/1"}],"TotalRecordCount":1}`,
			statusCode:   200,
			expectError:  false,
			expectMapLen: 7,
			description:  "Handles concurrent requests for 7 users (exceeds 5 semaphore limit)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			server := mockServer(tt.mockResponse, tt.statusCode)
			defer server.Close()

			// Create client
			client := NewClient(server.URL, "test-key", "admin-user")

			// Execute
			result, err := client.GetSeenMoviesForAllUsers(tt.users)

			// Validate error handling
			if tt.expectError {
				assert.Error(t, err, "Expected error: "+tt.description)
				assert.Nil(t, result, "Result should be nil on error")
			} else {
				assert.NoError(t, err, "Unexpected error: "+tt.description)
				if tt.expectMapLen == 0 {
					// For empty list, result should be empty map not nil
					assert.NotNil(t, result, "Result should be map (empty or not)")
					assert.Equal(t, tt.expectMapLen, len(result), "Expected map length: "+tt.description)
				} else {
					// For non-empty list, check all users are in the map
					assert.NotNil(t, result, "Result should not be nil: "+tt.description)
					assert.Equal(t, tt.expectMapLen, len(result), "Expected all users in map: "+tt.description)

					// Verify each user has an entry
					for _, user := range tt.users {
						_, exists := result[user.ID]
						assert.True(t, exists, "User "+user.ID+" should be in result map")

						// Each user should have at least one movie in the success case
						if tt.statusCode == 200 {
							assert.NotNil(t, result[user.ID], "Movies list should not be nil for user "+user.ID)
						}
					}
				}
			}
		})
	}
}

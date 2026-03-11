package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func mockServer(response string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(response))
	}))
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

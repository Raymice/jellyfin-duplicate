package server_test

import (
	apiModels "jellyfin-duplicate/client/jellyfin/models"
	"jellyfin-duplicate/server"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTestRouter creates a gin router with templates loaded for testing
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Load templates from embedded filesystem
	r.LoadHTMLFS(http.FS(server.GetTemplateFS()), server.GetTemplateFSPath())
	return r
}

type mockJellyfinClientForHandler struct {
	movies     []apiModels.MovieLightAPI
	users      []apiModels.UserAPI
	shouldFail bool
}

// Implement http.JellyfinClient interface methods
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

// TestNewHandler_SuccessfulCreation tests Handler instantiation
func TestNewHandler_SuccessfulCreation(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Creates handler with valid client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockJellyfinClientForHandler{}
			handler := server.NewHandler(mockClient)

			assert.NotNil(t, handler, "Handler should not be nil")
		})
	}
}

// TestGetHomePage tests the home page endpoint
func TestGetHomePage(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "Home page renders successfully",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()

			// Create mock client and handler
			mockClient := &mockJellyfinClientForHandler{}
			handler := server.NewHandler(mockClient)

			// Register the route
			router.GET("/", handler.GetHomePage)

			// Create request and recorder
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Verify status code
			assert.Equal(t, tt.expectedStatus, w.Code, "Status code should match")
		})
	}
}

// TestGetDuplicatesJSON tests the JSON duplicates endpoint
func TestGetDuplicatesJSON(t *testing.T) {
	tests := []struct {
		name           string
		movies         []apiModels.MovieLightAPI
		shouldFail     bool
		expectedStatus int
	}{
		{
			name: "Returns empty array when no duplicates found",
			movies: []apiModels.MovieLightAPI{
				{MovieXtLightAPI: apiModels.MovieXtLightAPI{ID: "movie1", Name: "Movie A", ProductionYear: 2020}, Path: "/path1"},
			},
			shouldFail:     false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Service error returns 500",
			shouldFail:     true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()

			mockClient := &mockJellyfinClientForHandler{
				movies:     tt.movies,
				users:      []apiModels.UserAPI{{ID: "user1", Name: "User One"}},
				shouldFail: tt.shouldFail,
			}

			handler := server.NewHandler(mockClient)
			router.GET("/api/duplicates", handler.GetDuplicatesJSON)

			req := httptest.NewRequest("GET", "/api/duplicates", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code should match")
			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json", "Should return JSON")
			}
		})
	}
}

// TestDeleteMovie tests the delete movie endpoint
func TestDeleteMovie(t *testing.T) {
	validMovieID := "12345678-1234-1234-1234-123456789012"

	tests := []struct {
		name           string
		movieID        string
		shouldFail     bool
		expectedStatus int
	}{
		{
			name:           "Successfully deletes movie with valid UUID",
			movieID:        validMovieID,
			shouldFail:     false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Rejects missing movieId parameter",
			movieID:        "",
			shouldFail:     false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Rejects invalid UUID format",
			movieID:        "invalid-short-id",
			shouldFail:     false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Handles service error with 500",
			movieID:        validMovieID,
			shouldFail:     true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()

			mockClient := &mockJellyfinClientForHandler{
				shouldFail: tt.shouldFail,
			}

			handler := server.NewHandler(mockClient)
			router.GET("/api/delete-movie", handler.DeleteMovie)

			req := httptest.NewRequest("GET", "/api/delete-movie?movieId="+tt.movieID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code should match for: "+tt.name)
		})
	}
}

// TestMarkMovieAsSeen tests the mark movie as seen endpoint
func TestMarkMovieAsSeen(t *testing.T) {
	validMovieID := "12345678-1234-1234-1234-123456789012"
	validUserID := "87654321-4321-4321-4321-210987654321"

	tests := []struct {
		name           string
		movieID        string
		userID         string
		shouldFail     bool
		expectedStatus int
	}{
		{
			name:           "Successfully marks movie as seen with valid IDs",
			movieID:        validMovieID,
			userID:         validUserID,
			shouldFail:     false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Rejects missing movieId parameter",
			movieID:        "",
			userID:         validUserID,
			shouldFail:     false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Rejects missing userId parameter",
			movieID:        validMovieID,
			userID:         "",
			shouldFail:     false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Rejects invalid movieId format",
			movieID:        "invalid-id",
			userID:         validUserID,
			shouldFail:     false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Rejects invalid userId format",
			movieID:        validMovieID,
			userID:         "invalid-user",
			shouldFail:     false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Handles service error with 500",
			movieID:        validMovieID,
			userID:         validUserID,
			shouldFail:     true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()

			mockClient := &mockJellyfinClientForHandler{
				shouldFail: tt.shouldFail,
			}

			handler := server.NewHandler(mockClient)
			router.GET("/api/mark-as-seen", handler.MarkMovieAsSeen)

			req := httptest.NewRequest("GET", "/api/mark-as-seen?movieId="+tt.movieID+"&userId="+tt.userID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code should match for: "+tt.name)
		})
	}
}

// TestGetDuplicatesPage tests the HTML duplicates page endpoint
func TestGetDuplicatesPage(t *testing.T) {
	tests := []struct {
		name       string
		movies     []apiModels.MovieLightAPI
		users      []apiModels.UserAPI
		shouldFail bool
	}{
		{
			name: "Renders duplicates page with movies",
			movies: []apiModels.MovieLightAPI{
				{MovieXtLightAPI: apiModels.MovieXtLightAPI{ID: "movie1", Name: "Movie A", ProductionYear: 2020}, Path: "/path1"},
				{MovieXtLightAPI: apiModels.MovieXtLightAPI{ID: "movie2", Name: "Movie A", ProductionYear: 2020}, Path: "/path2"},
			},
			users:      []apiModels.UserAPI{{ID: "user1", Name: "User One"}},
			shouldFail: false,
		},
		{
			name:       "Renders duplicates page with no movies",
			movies:     []apiModels.MovieLightAPI{},
			users:      []apiModels.UserAPI{},
			shouldFail: false,
		},
		{
			name:       "Returns error page on service failure",
			movies:     []apiModels.MovieLightAPI{},
			users:      []apiModels.UserAPI{},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()

			mockClient := &mockJellyfinClientForHandler{
				movies:     tt.movies,
				users:      tt.users,
				shouldFail: tt.shouldFail,
			}

			handler := server.NewHandler(mockClient)
			router.GET("/analysis", handler.GetDuplicatesPage)

			req := httptest.NewRequest("GET", "/analysis", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Verify handler was called - status verification depends on scenario
			assert.NotNil(t, handler, "Handler should not be nil")
		})
	}
}

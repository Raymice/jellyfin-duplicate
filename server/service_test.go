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

type fakeJellyfinClient struct {
	movies         []apiModels.MovieLightAPI
	users          []apiModels.UserAPI
	userSeenMovies map[string][]apiModels.MovieLightAPI
	movieUserData  map[string]map[string]apiModels.UserDataAPI // movieID -> userID -> UserDataAPI
	movieDetails   map[string]apiModels.MovieLightExtendedAPI  // movieID -> MovieLightExtendedAPI
	movieNames     map[string]string                           // movieID -> movieName
	userNames      map[string]string                           // userID -> userName
	deletedMovies  []string                                    // track deleted movies
	markedAsPlayed []string                                    // track marked as played calls
	shouldError    bool
	errorMessage   string
}

func (f *fakeJellyfinClient) GetAllMovies() ([]apiModels.MovieLightAPI, error) {
	if f.shouldError {
		return nil, fmt.Errorf("%s", f.errorMessage)
	}
	return f.movies, nil
}

func (f *fakeJellyfinClient) GetAllUsers() ([]apiModels.UserAPI, error) {
	if f.shouldError {
		return nil, fmt.Errorf("%s", f.errorMessage)
	}
	return f.users, nil
}

func (f *fakeJellyfinClient) GetSeenMoviesForAllUsers(users []apiModels.UserAPI) (map[string][]apiModels.MovieLightAPI, error) {
	if f.shouldError {
		return nil, fmt.Errorf("%s", f.errorMessage)
	}
	return f.userSeenMovies, nil
}

func (f *fakeJellyfinClient) GetMovieDetails(movieID string) (apiModels.MovieLightExtendedAPI, error) {
	if f.shouldError {
		return apiModels.MovieLightExtendedAPI{}, fmt.Errorf("%s", f.errorMessage)
	}
	if f.movieDetails == nil {
		panic("movieDetails not set up in fake client")
	}
	if details, ok := f.movieDetails[movieID]; ok {
		return details, nil
	}
	panic("movieID not found in fake client movieDetails")
}

func (f *fakeJellyfinClient) GetMovieUserData(movieID string, userID string) (apiModels.UserDataAPI, error) {
	if f.shouldError {
		return apiModels.UserDataAPI{}, fmt.Errorf("%s", f.errorMessage)
	}
	if f.movieUserData == nil {
		panic("movieUserData not set up in fake client")
	}
	if userMap, ok := f.movieUserData[movieID]; ok {
		if userData, ok := userMap[userID]; ok {
			return userData, nil
		}
	}
	panic("movieID or userID not found in fake client")
}

func (f *fakeJellyfinClient) GetMovieName(movieID string) (string, error) {
	if f.shouldError {
		return "", fmt.Errorf("%s", f.errorMessage)
	}
	if f.movieNames == nil {
		return "", fmt.Errorf("movie name not found")
	}
	if name, ok := f.movieNames[movieID]; ok {
		return name, nil
	}
	return "", fmt.Errorf("movie name not found")
}

func (f *fakeJellyfinClient) GetUserName(userID string) (string, error) {
	if f.shouldError {
		return "", fmt.Errorf("%s", f.errorMessage)
	}
	if f.userNames == nil {
		return "", fmt.Errorf("user name not found")
	}
	if name, ok := f.userNames[userID]; ok {
		return name, nil
	}
	return "", fmt.Errorf("user name not found")
}

func (f *fakeJellyfinClient) DeleteMovie(movieID string) error {
	if f.shouldError {
		return fmt.Errorf("%s", f.errorMessage)
	}
	if f.deletedMovies == nil {
		f.deletedMovies = []string{}
	}
	f.deletedMovies = append(f.deletedMovies, movieID)
	return nil
}

func (f *fakeJellyfinClient) MarkMovieAsPlayed(movieID string, userID string, movieName string, userName string) error {
	if f.shouldError {
		return fmt.Errorf("%s", f.errorMessage)
	}
	if f.markedAsPlayed == nil {
		f.markedAsPlayed = []string{}
	}
	f.markedAsPlayed = append(f.markedAsPlayed, fmt.Sprintf("%s:%s", movieID, userID))
	return nil
}

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

func TestServerService_GetMultiUserPlayStatus(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			allMovies := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/all_movies.json", functionName, useCase), []apiModels.MovieLightAPI{})
			users := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/users.json", functionName, useCase), []apiModels.UserAPI{})
			userSeenMovies := make(map[string][]apiModels.MovieLightAPI)
			for _, user := range users {
				userSeenMovies[user.ID] = test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/%s_seen_movies.json", functionName, useCase, user.ID), []apiModels.MovieLightAPI{})
			}
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), []serverModels.MovieLightStatusDTO{})

			service := server.NewService(&fakeJellyfinClient{movies: allMovies, users: users, userSeenMovies: userSeenMovies})
			got, gotErr := service.GetMultiUserPlayStatus()

			assert.NoError(t, gotErr)
			assert.ElementsMatch(t, expected, got)
		})
	}
}

func TestServerService_HasIdenticalPlayStatus(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			movie1 := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/movie1.json", functionName, useCase), serverModels.MovieLightStatusDTO{})
			movie2 := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/movie2.json", functionName, useCase), serverModels.MovieLightStatusDTO{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), false)
			// Execution
			got := server.NewService(&jellyfinClients.Client{}).HasIdenticalPlayStatus(movie1, movie2)
			// Validation
			assert.Equal(t, expected, got, "Should return the expected identical play status result")
		})
	}
}

func TestServerService_GetUserPlayStatus(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			movieID := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				MovieID  string                `json:"movieId"`
				UserID   string                `json:"userId"`
				UserData apiModels.UserDataAPI `json:"userData"`
			}{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), serverModels.UserPlayStatusDTO{})

			// Setup fake client with user data
			fakeClient := &fakeJellyfinClient{
				movieUserData: map[string]map[string]apiModels.UserDataAPI{
					movieID.MovieID: {
						movieID.UserID: movieID.UserData,
					},
				},
			}

			// Execution
			got, gotErr := server.NewService(fakeClient).GetUserPlayStatus(movieID.MovieID, movieID.UserID)

			// Validation
			assert.NoError(t, gotErr)
			assert.Equal(t, expected, got, "Should return the expected user play status")
		})
	}
}

func TestServerService_GetPlayStatusForAllUsers(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			dup := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), serverModels.DuplicateResultDTO{})
			users := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/users.json", functionName, useCase), []apiModels.UserAPI{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), serverModels.DuplicateResultDTO{})

			// Setup fake client with users and user data for both movies
			movieUserData := make(map[string]map[string]apiModels.UserDataAPI)

			for _, user := range users {
				// Load user data for movie 1
				movie1UserData := test.ParseFromJsonFile(t,
					fmt.Sprintf("%s/%s/%s_movie1_user_data.json", functionName, useCase, user.ID),
					apiModels.UserDataAPI{})

				// Load user data for movie 2
				movie2UserData := test.ParseFromJsonFile(t,
					fmt.Sprintf("%s/%s/%s_movie2_user_data.json", functionName, useCase, user.ID),
					apiModels.UserDataAPI{})

				if _, ok := movieUserData[dup.Movie1.ID]; !ok {
					movieUserData[dup.Movie1.ID] = make(map[string]apiModels.UserDataAPI)
				}
				movieUserData[dup.Movie1.ID][user.ID] = movie1UserData

				if _, ok := movieUserData[dup.Movie2.ID]; !ok {
					movieUserData[dup.Movie2.ID] = make(map[string]apiModels.UserDataAPI)
				}
				movieUserData[dup.Movie2.ID][user.ID] = movie2UserData
			}

			fakeClient := &fakeJellyfinClient{
				users:         users,
				movieUserData: movieUserData,
			}

			// Execution
			got, gotErr := server.NewService(fakeClient).GetPlayStatusForAllUsers(dup)

			// Validation
			assert.NoError(t, gotErr)
			assert.Equal(t, expected.Movie1.UserPlayStatuses, got.Movie1.UserPlayStatuses, "Movie1 play statuses should match")
			assert.Equal(t, expected.Movie2.UserPlayStatuses, got.Movie2.UserPlayStatuses, "Movie2 play statuses should match")
		})
	}
}

func TestServerService_FindDuplicates(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Load fixture data
			allMovies := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/all_movies.json", functionName, useCase), []apiModels.MovieLightAPI{})
			users := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/users.json", functionName, useCase), []apiModels.UserAPI{})
			movieDetails := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/movie_details.json", functionName, useCase), map[string]apiModels.MovieLightExtendedAPI{})

			// Load user play data
			userSeenMovies := make(map[string][]apiModels.MovieLightAPI)
			movieUserData := make(map[string]map[string]apiModels.UserDataAPI)

			for _, user := range users {
				userSeenMovies[user.ID] = test.ParseFromJsonFile(t,
					fmt.Sprintf("%s/%s/%s_seen_movies.json", functionName, useCase, user.ID),
					[]apiModels.MovieLightAPI{})

				for _, movie := range allMovies {
					if _, ok := movieUserData[movie.ID]; !ok {
						movieUserData[movie.ID] = make(map[string]apiModels.UserDataAPI)
					}
					userData := test.ParseFromJsonFile(t,
						fmt.Sprintf("%s/%s/%s_%s_user_data.json", functionName, useCase, user.ID, movie.ID),
						apiModels.UserDataAPI{})
					movieUserData[movie.ID][user.ID] = userData
				}
			}

			// Setup fake client
			fakeClient := &fakeJellyfinClient{
				movies:         allMovies,
				users:          users,
				userSeenMovies: userSeenMovies,
				movieUserData:  movieUserData,
				movieDetails:   movieDetails,
			}

			// Execution
			got, gotErr := server.NewService(fakeClient).FindDuplicates()

			// Validation
			assert.NoError(t, gotErr)
			// Verify we got at least the expected number of duplicates
			assert.True(t, len(got) > 0, "Should find at least one duplicate pair")
			// Verify first duplicate has the right movies
			assert.True(t, (got[0].Movie1.ID == "movie1" && got[0].Movie2.ID == "movie2") || (got[0].Movie1.ID == "movie2" && got[0].Movie2.ID == "movie1"),
				"Should match the expected movie pair IDs")
		})
	}
}

func TestServerService_DeleteMovie(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Load input data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				MovieID      string `json:"movieId"`
				ShouldError  bool   `json:"shouldError"`
				ErrorMessage string `json:"errorMessage"`
			}{})

			fakeClient := &fakeJellyfinClient{
				shouldError:  input.ShouldError,
				errorMessage: input.ErrorMessage,
			}

			// Execution
			err := server.NewService(fakeClient).DeleteMovie(input.MovieID)

			// Validation
			if input.ShouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, fakeClient.deletedMovies, input.MovieID, "Movie should be marked as deleted")
			}
		})
	}
}

func TestServerService_MarkMovieAsSeen(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Load input data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				MovieID      string `json:"movieId"`
				UserID       string `json:"userId"`
				ShouldError  bool   `json:"shouldError"`
				ErrorMessage string `json:"errorMessage"`
			}{})

			fakeClient := &fakeJellyfinClient{
				movieNames: map[string]string{
					input.MovieID: "Test Movie",
				},
				userNames: map[string]string{
					input.UserID: "Test User",
				},
				shouldError:  input.ShouldError,
				errorMessage: input.ErrorMessage,
			}

			// Execution
			err := server.NewService(fakeClient).MarkMovieAsSeen(input.MovieID, input.UserID)

			// Validation
			if input.ShouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				expectedCall := fmt.Sprintf("%s:%s", input.MovieID, input.UserID)
				assert.Contains(t, fakeClient.markedAsPlayed, expectedCall, "Movie should be marked as played")
			}
		})
	}
}

func TestIsUUIDFormtatted(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			input := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/input.json", functionName, useCase), struct {
				ID string `json:"id"`
			}{})
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), struct {
				IsValidUUID bool `json:"isValidUUID"`
			}{})

			// Execution
			got := server.IsUUIDFormtatted(input.ID)

			// Validation
			assert.Equal(t, expected.IsValidUUID, got, "Should return whether the string is a valid UUID format")
		})
	}
}

func TestNewService(t *testing.T) {
	functionName := test.GetFuncName()
	useCases := test.GetTestUseCases(functionName)

	for _, useCase := range useCases {
		t.Run(useCase, func(t *testing.T) {
			// Data
			expected := test.ParseFromJsonFile(t, fmt.Sprintf("%s/%s/expected.json", functionName, useCase), struct {
				ServiceNotNil bool `json:"serviceNotNil"`
			}{})

			// Execution
			fakeClient := &fakeJellyfinClient{}
			service := server.NewService(fakeClient)

			// Validation
			if expected.ServiceNotNil {
				assert.NotNil(t, service, "Service should not be nil")
			}
		})
	}
}

func (f *fakeJellyfinClient) GetSeenMoviesForUser(userID string) ([]apiModels.MovieLightAPI, error) {
	if f.shouldError {
		return nil, fmt.Errorf("%s", f.errorMessage)
	}
	if f.userSeenMovies == nil {
		return []apiModels.MovieLightAPI{}, nil
	}
	return f.userSeenMovies[userID], nil
}

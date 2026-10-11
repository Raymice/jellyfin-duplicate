package http

import (
	"context"
	"encoding/json"
	"fmt"
	"jellyfin-duplicate/client/jellyfin/models"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// HTTP Status Code Descriptions for better error messages
var httpStatusDescriptions = map[int]string{
	200: "OK",
	201: "Created",
	204: "No Content",
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	408: "Request Timeout",
	500: "Internal Server Error",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
}

type Client struct {
	legacyAuth bool
	baseURL    string
	apiKey     string
	userID     string
	client     *resty.Client
	userCache  map[string]string // userID -> userName cache
	cacheMutex sync.Mutex        // mutex to protect cache access
}

type JellyfinClient interface {
	GetAllMovies() ([]models.MovieLightAPI, error)
	GetAllUsers() ([]models.UserAPI, error)
	GetSeenMoviesForAllUsers(users []models.UserAPI) (map[string][]models.MovieLightAPI, error)
	GetMovieDetails(movieID string) (models.MovieLightExtendedAPI, error)
	GetMovieUserData(movieID string, userID string) (models.UserDataAPI, error)
	GetMovieName(movieID string) (string, error)
	GetUserName(userID string) (string, error)
	DeleteMovie(movieID string) error
	MarkMovieAsPlayed(movieID string, userID string, movieName string, userName string) error
}

func NewClient(baseURL, apiKey string, userID string, legacyAuth ...bool) *Client {
	return &Client{
		legacyAuth: len(legacyAuth) > 0 && legacyAuth[0],
		baseURL:    baseURL,
		apiKey:     apiKey,
		userID:     userID,
		client:     resty.New(),
		userCache:  make(map[string]string),
	}
}

// newRequest selects the configured authentication scheme for every API call.
func (c *Client) newRequest() *resty.Request {
	if c.legacyAuth {
		return c.client.R().SetHeader("X-MediaBrowser-Token", c.apiKey)
	}
	return c.client.R().SetHeader("Authorization", fmt.Sprintf(`MediaBrowser Token=%q`, c.apiKey))
}

// ConfigureAuthentication runs once at startup, before authenticated API calls.
// If detection fails, AUTO falls back to LATEST and explicit modes remain selected.
func (c *Client) ConfigureAuthentication(mode string) error {
	mode = strings.ToUpper(strings.TrimSpace(mode))
	if mode == "" {
		mode = "AUTO"
	}
	if mode != "AUTO" && mode != "LEGACY" && mode != "LATEST" {
		return fmt.Errorf("JELLYFIN_AUTH must be AUTO, LEGACY, or LATEST")
	}
	c.legacyAuth = mode == "LEGACY"
	defer func() {
		if c.legacyAuth {
			logrus.Infof("Jellyfin authentication: LEGACY (X-MediaBrowser-Token header; configured mode: %s)", mode)
		} else {
			logrus.Infof("Jellyfin authentication: LATEST (Authorization: MediaBrowser header; configured mode: %s)", mode)
		}
	}()
	version, err := c.GetServerVersion()
	if err != nil {
		logrus.Info("Jellyfin server version: unknown (public version request failed)")
		return err
	}
	logrus.Infof("Jellyfin server version: %s", version)
	parts := regexp.MustCompile(`^(\d+)\.\d+\.\d+(?:[-+].+)?$`).FindStringSubmatch(version)
	if parts == nil {
		return fmt.Errorf("unrecognized Jellyfin server version %q", version)
	}
	major, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid Jellyfin major version: %w", err)
	}
	useLegacy := major < 12
	if mode == "AUTO" {
		c.legacyAuth = useLegacy
	} else if c.legacyAuth != useLegacy {
		recommended := "LATEST"
		if useLegacy {
			recommended = "LEGACY"
		}
		logrus.Warnf("JELLYFIN_AUTH=%s differs from the recommended %s authentication for Jellyfin %s", mode, recommended, version)
	}
	return nil
}

// GetServerVersion queries public system information without authentication.
func (c *Client) GetServerVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := c.client.R().SetContext(ctx).
		Get(strings.TrimRight(c.baseURL, "/") + "/System/Info/Public")
	if err != nil {
		return "", fmt.Errorf("failed to fetch Jellyfin server version: %w", err)
	}
	if err := checkHTTPResponse(resp, 200); err != nil {
		return "", err
	}
	var info struct {
		Version string `json:"Version"`
	}
	if err := json.Unmarshal(resp.Body(), &info); err != nil {
		return "", fmt.Errorf("failed to parse Jellyfin public system information: %w", err)
	}
	if strings.TrimSpace(info.Version) == "" {
		return "", fmt.Errorf("Jellyfin public system information contains no version")
	}
	return info.Version, nil
}

func (c *Client) GetAllMovies() ([]models.MovieLightAPI, error) {
	logrus.Info("Fetching all movies from Jellyfin in parallel...")
	var movies []models.MovieLightAPI

	// Get all libraries first
	logrus.Debug("Getting libraries...")
	libraries, err := c.GetLibraries()
	if err != nil {
		return nil, fmt.Errorf("failed to get libraries: %v", err)
	}
	logrus.Infof("Found %d libraries", len(libraries))

	// Use channels for parallel fetching
	movieChannel := make(chan []models.MovieLightAPI, len(libraries))
	errorChannel := make(chan error, len(libraries))
	var wg sync.WaitGroup

	// Limit concurrent goroutines to avoid overwhelming the system
	// This prevents too many simultaneous API calls
	maxConcurrent := 5
	semaphore := make(chan struct{}, maxConcurrent)

	// For each library, get movies in parallel
	for _, library := range libraries {
		wg.Add(1)
		go func(lib models.LibraryAPI) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			logrus.Debugf("Fetching movies from library: %s", lib.Name)
			libraryMovies, err := c.getMoviesFromLibrary(lib.ID)
			if err != nil {
				errorChannel <- fmt.Errorf("failed to get movies from library %s: %v", lib.Name, err)
				return
			}
			logrus.Infof("Found %d movies in library: %s", len(libraryMovies), lib.Name)
			movieChannel <- libraryMovies
		}(library)
	}

	// Close channels when all goroutines are done
	go func() {
		wg.Wait()
		close(movieChannel)
		close(errorChannel)
	}()

	// Collect results from channels
	for libraryMovies := range movieChannel {
		movies = append(movies, libraryMovies...)
	}

	// Check for any errors
	if len(errorChannel) > 0 {
		var errorMessages []string
		for err := range errorChannel {
			errorMessages = append(errorMessages, err.Error())
		}
		return nil, fmt.Errorf("errors occurred while fetching movies: %s", strings.Join(errorMessages, "; "))
	}

	logrus.Infof("Total movies fetched: %d", len(movies))
	return movies, nil
}

func (c *Client) GetLibraries() ([]models.LibraryAPI, error) {
	if c.userID == "" {
		return nil, fmt.Errorf("user ID not set")
	}

	start := time.Now()

	resp, err := c.newRequest().
		Get(fmt.Sprintf("%s/Users/%s/Views", c.baseURL, c.userID))

	if err != nil {
		return nil, fmt.Errorf("failed to call Jellyfin API for libraries: %v", err)
	}

	// Check HTTP status code using our helper function
	err = checkHTTPResponse(resp, 200)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch libraries: %v", err)
	}

	// Parse the JSON response manually
	var result struct {
		Items []models.LibraryAPI `json:"Items"`
	}

	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Jellyfin API response: %v", err)
	}

	// Check if the response contains valid data
	if result.Items == nil {
		logrus.Errorf("Jellyfin API returned empty or invalid library data")
		return nil, fmt.Errorf("Jellyfin API returned empty or invalid library data")
	}

	logrus.Debugf("Successfully fetched %d libraries from Jellyfin in %v", len(result.Items), time.Since(start))
	return result.Items, nil
}

func (c *Client) getMoviesFromLibrary(libraryID string) ([]models.MovieLightAPI, error) {
	var allMovies []models.MovieLightAPI

	start := time.Now()

	// Start with the first page
	startIndex := 0
	limit := 2000

	for {
		var result struct {
			Items            []models.MovieLightAPI `json:"Items"`
			TotalRecordCount int                    `json:"TotalRecordCount"`
		}

		resp, err := c.newRequest().
			SetQueryParam("recursive", "true").
			SetQueryParam("includeItemTypes", "Movie").
			SetQueryParam("parentId", libraryID).
			SetQueryParam("enableUserData", "false").
			SetQueryParam("enableImages", "false").
			SetQueryParam("enableTotalRecordCount", "true").
			SetQueryParam("fields", "Path").
			SetQueryParam("startIndex", fmt.Sprintf("%d", startIndex)).
			SetQueryParam("limit", fmt.Sprintf("%d", limit)).
			SetResult(&result).
			Get(fmt.Sprintf("%s/Items", c.baseURL))

		if err != nil {
			return nil, fmt.Errorf("failed to call Jellyfin API for movies in library %s: %v", libraryID, err)
		}

		// Check HTTP status code
		err = checkHTTPResponse(resp, 200)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch movies from library %s: %v", libraryID, err)
		}

		// Add movies from this page to our collection
		allMovies = append(allMovies, result.Items...)

		// Check if we've fetched all movies
		if len(allMovies) >= result.TotalRecordCount {
			break
		}

		// Move to the next page
		startIndex += limit
	}

	logrus.Infof("Finished fetching movies from library %s in %v", libraryID, time.Since(start))
	return allMovies, nil
}

// GetAllUsers fetches all users from Jellyfin and populates the user cache
func (c *Client) GetAllUsers() ([]models.UserAPI, error) {

	var users []models.UserAPI

	start := time.Now()
	logrus.Info("Fetching all users from Jellyfin...")

	resp, err := c.newRequest().
		SetResult(&users).
		Get(fmt.Sprintf("%s/Users", c.baseURL))

	if err != nil {
		return nil, fmt.Errorf("failed to call Jellyfin API for users: %v", err)
	}

	// Check HTTP status code
	err = checkHTTPResponse(resp, 200)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %v", err)
	}

	// Populate user cache with all fetched users
	c.cacheMutex.Lock()
	for _, user := range users {
		c.userCache[user.ID] = user.Name
	}
	c.cacheMutex.Unlock()

	logrus.Infof("Found %d users and populated user cache in %v", len(users), time.Since(start))
	return users, nil
}

func (c *Client) GetMovieUserData(movieID string, userID string) (models.UserDataAPI, error) {
	var result models.UserDataAPI

	resp, err := c.newRequest().
		SetQueryParam("userId", userID).
		SetResult(&result).
		Get(fmt.Sprintf("%s/UserItems/%s/UserData", c.baseURL, movieID))

	if err != nil {
		return models.UserDataAPI{}, fmt.Errorf("failed to call Jellyfin API for user play status: %v", err)
	}

	// Check HTTP status code
	err = checkHTTPResponse(resp, 200)
	if err != nil {
		return models.UserDataAPI{}, fmt.Errorf("failed to fetch user play status: %v", err)
	}

	return result, nil
}

func (c *Client) GetMovieDetails(movieID string) (models.MovieLightExtendedAPI, error) {
	var result models.MovieLightExtendedAPI

	resp, err := c.newRequest().
		SetQueryParam("userId", c.userID).
		SetQueryParam("enableImages", "false").
		SetQueryParam("enableUserData", "false").
		SetQueryParam("fields", "Path").
		SetResult(&result).
		Get(fmt.Sprintf("%s/Items/%s", c.baseURL, movieID))

	if err != nil {
		return models.MovieLightExtendedAPI{}, fmt.Errorf("failed to call Jellyfin API for movie details: %v", err)
	}

	// Check HTTP status code
	err = checkHTTPResponse(resp, 200)
	if err != nil {
		return models.MovieLightExtendedAPI{}, fmt.Errorf("failed to fetch movie details: %v", err)
	}

	return result, nil
}

// GetSeenMoviesForUser fetches all movies that a specific user has seen (played)
func (c *Client) GetSeenMoviesForUser(userID string) ([]models.MovieLightAPI, error) {
	var allMovies []models.MovieLightAPI

	// Start with the first page
	startIndex := 0
	limit := 2000

	for {
		var result struct {
			Items            []models.MovieLightAPI `json:"Items"`
			TotalRecordCount int                    `json:"TotalRecordCount"`
		}

		resp, err := c.newRequest().
			SetQueryParam("recursive", "true").
			SetQueryParam("includeItemTypes", "Movie").
			SetQueryParam("isPlayed", "true").
			SetQueryParam("userId", userID).
			SetQueryParam("enableUserData", "false").
			SetQueryParam("enableImages", "false").
			SetQueryParam("startIndex", fmt.Sprintf("%d", startIndex)).
			SetQueryParam("limit", fmt.Sprintf("%d", limit)).
			SetResult(&result).
			Get(fmt.Sprintf("%s/Items", c.baseURL))

		if err != nil {
			return nil, fmt.Errorf("failed to fetch seen movies for user %s: %v", userID, err)
		}

		// Debug: Log the raw response if there's an issue
		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("API request failed with status %d for user %s", resp.StatusCode(), userID)
		}

		// Add movies from this page to our collection
		allMovies = append(allMovies, result.Items...)

		// Check if we've fetched all movies
		if len(allMovies) >= result.TotalRecordCount {
			break
		}

		// Move to the next page
		startIndex += limit
	}

	return allMovies, nil
}

// GetSeenMoviesForAllUsers fetches seen movies for all users in parallel (max 5 concurrent)
func (c *Client) GetSeenMoviesForAllUsers(users []models.UserAPI) (map[string][]models.MovieLightAPI, error) {
	userSeenMovies := make(map[string][]models.MovieLightAPI)
	var mu sync.Mutex
	var wg sync.WaitGroup

	start := time.Now()
	logrus.Infof("Fetching seen movies for %d users in parallel...", len(users))

	// Limit concurrent goroutines to 5
	semaphore := make(chan struct{}, 5)

	var errors []error

	for _, user := range users {
		wg.Add(1)
		go func(u models.UserAPI) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			logrus.Debugf("Fetching seen movies for user: %s", u.Name)
			seenMovies, err := c.GetSeenMoviesForUser(u.ID)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to get seen movies for user %s: %v", u.Name, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			userSeenMovies[u.ID] = seenMovies
			mu.Unlock()
			logrus.Infof("Found %d seen movies for user: %s", len(seenMovies), u.Name)
		}(user)
	}

	wg.Wait()

	if len(errors) > 0 {
		return nil, fmt.Errorf("errors occurred while fetching seen movies: %v", errors)
	}

	logrus.Infof("Successfully fetched seen movies for all %d users in %v", len(users), time.Since(start))
	return userSeenMovies, nil
}

// GetMovieName gets the name of a movie by its ID
func (c *Client) GetMovieName(movieID string) (string, error) {
	// Use the user-specific items endpoint which returns more complete data
	if c.userID == "" {
		return "", fmt.Errorf("user ID not set for movie name lookup")
	}

	var result struct {
		Name string `json:"Name"`
	}

	resp, err := c.newRequest().
		SetQueryParam("Fields", "Path").
		SetResult(&result).
		Get(fmt.Sprintf("%s/Users/%s/Items/%s", c.baseURL, c.userID, movieID))

	if err != nil {
		return "", fmt.Errorf("failed to call Jellyfin API for movie name: %v", err)
	}

	// Check HTTP status code
	err = checkHTTPResponse(resp, 200)
	if err != nil {
		return "", fmt.Errorf("failed to fetch movie name: %v", err)
	}

	// Fallback: if Name is empty, try the basic Items endpoint
	if result.Name == "" {
		var basicResult struct {
			Name string `json:"Name"`
		}
		resp, err := c.newRequest().
			SetResult(&basicResult).
			Get(fmt.Sprintf("%s/Items/%s", c.baseURL, movieID))

		if err != nil {
			return "", fmt.Errorf("failed to call Jellyfin API for movie name (fallback): %v", err)
		}

		// Check HTTP status code for fallback
		err = checkHTTPResponse(resp, 200)
		if err != nil {
			return "", fmt.Errorf("failed to fetch movie name (fallback): %v", err)
		}

		return basicResult.Name, nil
	}

	return result.Name, nil
}

// GetUserName gets the name of a user by their ID with caching
func (c *Client) GetUserName(userID string) (string, error) {
	// Check cache first
	c.cacheMutex.Lock()
	if cachedName, exists := c.userCache[userID]; exists {
		c.cacheMutex.Unlock()
		return cachedName, nil
	}
	c.cacheMutex.Unlock()

	// Cache miss, fetch from API
	var result struct {
		Name string `json:"Name"`
	}

	resp, err := c.newRequest().
		SetResult(&result).
		Get(fmt.Sprintf("%s/Users/%s", c.baseURL, userID))

	if err != nil {
		return "", fmt.Errorf("failed to call Jellyfin API for user name: %v", err)
	}

	// Check HTTP status code
	err = checkHTTPResponse(resp, 200)
	if err != nil {
		return "", fmt.Errorf("failed to fetch user name: %v", err)
	}

	// Cache the result
	c.cacheMutex.Lock()
	c.userCache[userID] = result.Name
	c.cacheMutex.Unlock()

	return result.Name, nil
}

// MarkMovieAsPlayed marks a movie as played for a specific user using Jellyfin API
func (c *Client) MarkMovieAsPlayed(movieID string, userID string, movieName string, userName string) error {
	logrus.Infof("Marking movie %s (%s) as played for user %s (%s)", movieName, movieID, userName, userID)

	// Jellyfin API endpoint to mark an item as played
	// Alternative endpoint format that might work better
	url := fmt.Sprintf("%s/Users/%s/PlayedItems/%s", c.baseURL, userID, movieID)
	logrus.Debugf("Using URL: %s", url)

	resp, err := c.newRequest().
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		logrus.Errorf("Network error marking movie as played: %v", err)
		return fmt.Errorf("failed to mark movie as played: %v", err)
	}

	// Check response status code
	statusCode := resp.StatusCode()
	logrus.Debugf("Jellyfin API response status: %d", statusCode)

	// Debug: Log the full response for troubleshooting
	if statusCode != 204 && statusCode != 200 {
		logrus.Warnf("Response body: %s", string(resp.Body()))
	}

	// Jellyfin API returns 204 No Content on success for this endpoint
	// Some versions might return 200 OK
	if statusCode != 204 && statusCode != 200 {
		logrus.Errorf("Unexpected status code %d when marking movie as played", statusCode)
		return fmt.Errorf("unexpected status code %d when marking movie as played", statusCode)
	}

	logrus.Infof("Successfully marked movie %s (%s) as played for user %s (%s)", movieName, movieID, userName, userID)
	return nil
}

// DeleteMovie deletes a movie from Jellyfin using the API
func (c *Client) DeleteMovie(movieID string) error {
	logrus.Infof("Deleting movie %s from Jellyfin", movieID)

	// Jellyfin API endpoint to delete an item
	url := fmt.Sprintf("%s/Items/%s", c.baseURL, movieID)
	logrus.Debugf("Using delete URL: %s", url)

	resp, err := c.newRequest().
		Delete(url)

	if err != nil {
		logrus.Errorf("Network error deleting movie: %v", err)
		return fmt.Errorf("failed to delete movie: %v", err)
	}

	// Check response status code
	statusCode := resp.StatusCode()
	logrus.Debugf("Jellyfin API delete response status: %d", statusCode)

	// Debug: Log the full response for troubleshooting
	if statusCode != 204 && statusCode != 200 {
		logrus.Warnf("Delete response body: %s", string(resp.Body()))
	}

	// Jellyfin API returns 204 No Content on successful deletion
	// Some versions might return 200 OK
	if statusCode != 204 && statusCode != 200 {
		logrus.Errorf("Unexpected status code %d when deleting movie", statusCode)
		return fmt.Errorf("unexpected status code %d when deleting movie", statusCode)
	}

	logrus.Infof("Successfully deleted movie %s from Jellyfin", movieID)
	return nil
}

// checkHTTPResponse checks the HTTP response status code and returns an error if not successful
func checkHTTPResponse(resp *resty.Response, expectedStatusCodes ...int) error {
	statusCode := resp.StatusCode()

	// Check if status code is in the expected list
	for _, expectedCode := range expectedStatusCodes {
		if statusCode == expectedCode {
			return nil // Success
		}
	}

	// Get status description
	description := httpStatusDescriptions[statusCode]
	if description == "" {
		description = "Unknown Status"
	}

	// Log the error with response details
	logrus.Errorf("HTTP request failed with status %d (%s)", statusCode, description)
	logrus.Debugf("Response body: %s", string(resp.Body()))

	return fmt.Errorf("HTTP request failed with status %d (%s)", statusCode, description)
}

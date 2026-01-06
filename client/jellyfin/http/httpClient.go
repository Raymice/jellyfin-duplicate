package http

import (
	"fmt"
	"jellyfin-duplicate/client/jellyfin/models"
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type Client struct {
	baseURL    string
	apiKey     string
	userID     string
	client     *resty.Client
	userCache  map[string]string // userID -> userName cache
	cacheMutex sync.Mutex        // mutex to protect cache access
}

func NewClient(baseURL, apiKey string, userID string) *Client {
	return &Client{
		baseURL:   baseURL,
		apiKey:    apiKey,
		userID:    userID,
		client:    resty.New(),
		userCache: make(map[string]string),
	}
}

func (c *Client) GetAllMovies() ([]models.Movie, error) {
	logrus.Info("Fetching all movies from Jellyfin in parallel...")
	var movies []models.Movie

	// Get all libraries first
	logrus.Debug("Getting libraries...")
	libraries, err := c.getLibraries()
	if err != nil {
		return nil, fmt.Errorf("failed to get libraries: %v", err)
	}
	logrus.Infof("Found %d libraries", len(libraries))

	// Use channels for parallel fetching
	movieChannel := make(chan []models.Movie, len(libraries))
	errorChannel := make(chan error, len(libraries))
	var wg sync.WaitGroup

	// Limit concurrent goroutines to avoid overwhelming the system
	// This prevents too many simultaneous API calls
	maxConcurrent := 5
	semaphore := make(chan struct{}, maxConcurrent)

	// For each library, get movies in parallel
	for _, library := range libraries {
		wg.Add(1)
		go func(lib models.Library) {
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

func (c *Client) getLibraries() ([]models.Library, error) {
	if c.userID == "" {
		return nil, fmt.Errorf("user ID not set")
	}

	var result struct {
		Items []models.Library `json:"Items"`
	}

	_, err := c.client.R().
		SetHeader("X-MediaBrowser-Token", c.apiKey).
		SetResult(&result).
		Get(fmt.Sprintf("%s/Users/%s/Views", c.baseURL, c.userID))

	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

func (c *Client) getMoviesFromLibrary(libraryID string) ([]models.Movie, error) {
	var allMovies []models.Movie

	// Start with the first page
	startIndex := 0
	limit := 100 // Jellyfin's default limit, can be adjusted

	for {
		var result struct {
			Items            []models.Movie `json:"Items"`
			TotalRecordCount int            `json:"TotalRecordCount"`
		}

		_, err := c.client.R().
			SetHeader("X-MediaBrowser-Token", c.apiKey).
			SetQueryParam("Recursive", "true").
			SetQueryParam("IncludeItemTypes", "Movie").
			SetQueryParam("Fields", "ProviderIds,ProductionYear,Path,UserData").
			SetQueryParam("ParentId", libraryID).
			SetQueryParam("StartIndex", fmt.Sprintf("%d", startIndex)).
			SetQueryParam("Limit", fmt.Sprintf("%d", limit)).
			SetResult(&result).
			Get(fmt.Sprintf("%s/Items", c.baseURL))

		if err != nil {
			return nil, err
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

// GetUserPlayStatus fetches play status for a specific movie and user
// GetAllUsers fetches all users from Jellyfin and populates the user cache
func (c *Client) GetAllUsers() ([]models.User, error) {
	logrus.Info("Fetching all users from Jellyfin...")
	var users []models.User

	_, err := c.client.R().
		SetHeader("X-MediaBrowser-Token", c.apiKey).
		SetResult(&users).
		Get(fmt.Sprintf("%s/Users", c.baseURL))

	if err != nil {
		return nil, err
	}

	// Populate user cache with all fetched users
	c.cacheMutex.Lock()
	for _, user := range users {
		c.userCache[user.ID] = user.Name
	}
	c.cacheMutex.Unlock()

	logrus.Infof("Found %d users and populated user cache", len(users))
	return users, nil
}

func (c *Client) GetUserPlayStatus(movieID string, userID string) (models.UserPlayStatus, error) {
	var result struct {
		UserData struct {
			Played                bool   `json:"Played"`
			PlaybackPositionTicks int64  `json:"PlaybackPositionTicks"`
			PlayCount             int    `json:"PlayCount"`
			LastPlayedDate        string `json:"LastPlayedDate"`
		} `json:"UserData"`
	}

	_, err := c.client.R().
		SetHeader("X-MediaBrowser-Token", c.apiKey).
		SetResult(&result).
		Get(fmt.Sprintf("%s/Users/%s/Items/%s", c.baseURL, userID, movieID))

	if err != nil {
		return models.UserPlayStatus{}, err
	}

	return models.UserPlayStatus{
		UserID:    c.userID,
		UserName:  "Current User", // Would need to fetch user info separately
		Played:    result.UserData.Played,
		PlayCount: result.UserData.PlayCount,
	}, nil
}

// GetSeenMoviesForUser fetches all movies that a specific user has seen (played)
func (c *Client) GetSeenMoviesForUser(userID string) ([]models.Movie, error) {
	var allMovies []models.Movie

	// Start with the first page
	startIndex := 0
	limit := 100 // Jellyfin's default limit, can be adjusted

	for {
		var result struct {
			Items            []models.Movie `json:"Items"`
			TotalRecordCount int            `json:"TotalRecordCount"`
		}

		resp, err := c.client.R().
			SetHeader("X-MediaBrowser-Token", c.apiKey).
			SetQueryParam("Recursive", "true").
			SetQueryParam("IncludeItemTypes", "Movie").
			SetQueryParam("Fields", "ProviderIds,ProductionYear,Path,UserData").
			SetQueryParam("Filters", "IsPlayed").
			SetQueryParam("UserId", userID).
			SetQueryParam("StartIndex", fmt.Sprintf("%d", startIndex)).
			SetQueryParam("Limit", fmt.Sprintf("%d", limit)).
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
func (c *Client) GetSeenMoviesForAllUsers(users []models.User) (map[string][]models.Movie, error) {
	logrus.Infof("Fetching seen movies for %d users in parallel...", len(users))
	userSeenMovies := make(map[string][]models.Movie)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limit concurrent goroutines to 5
	semaphore := make(chan struct{}, 5)

	var errors []error

	for _, user := range users {
		wg.Add(1)
		go func(u models.User) {
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

	logrus.Infof("Successfully fetched seen movies for all %d users", len(users))
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

	_, err := c.client.R().
		SetHeader("X-MediaBrowser-Token", c.apiKey).
		SetQueryParam("Fields", "ProviderIds,ProductionYear,Path,UserData").
		SetResult(&result).
		Get(fmt.Sprintf("%s/Users/%s/Items/%s", c.baseURL, c.userID, movieID))

	if err != nil {
		return "", err
	}

	// Fallback: if Name is empty, try the basic Items endpoint
	if result.Name == "" {
		var basicResult struct {
			Name string `json:"Name"`
		}
		_, err := c.client.R().
			SetHeader("X-MediaBrowser-Token", c.apiKey).
			SetResult(&basicResult).
			Get(fmt.Sprintf("%s/Items/%s", c.baseURL, movieID))

		if err != nil {
			return "", err
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

	_, err := c.client.R().
		SetHeader("X-MediaBrowser-Token", c.apiKey).
		SetResult(&result).
		Get(fmt.Sprintf("%s/Users/%s", c.baseURL, userID))

	if err != nil {
		return "", err
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

	resp, err := c.client.R().
		SetHeader("X-MediaBrowser-Token", c.apiKey).
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

	resp, err := c.client.R().
		SetHeader("X-MediaBrowser-Token", c.apiKey).
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

// ReconcilePlayStatusWithAllMovies reconciles seen movies with all movies to create play status
func (c *Client) ReconcilePlayStatusWithAllMovies(allMovies []models.Movie, userSeenMovies map[string][]models.Movie, users []models.User) ([]models.Movie, error) {
	// Create a map of all movies by ID for quick lookup
	movieMap := make(map[string]models.Movie)
	for _, movie := range allMovies {
		movieMap[movie.ID] = movie
	}

	// For each user, mark their seen movies
	for _, user := range users {
		seenMovies, ok := userSeenMovies[user.ID]
		if !ok {
			// User has no seen movies, mark all movies as not seen
			for movieID, movie := range movieMap {
				playStatus := models.UserPlayStatus{
					UserID:   user.ID,
					UserName: user.Name,
					Played:   false,
				}
				movie.UserPlayStatuses = append(movie.UserPlayStatuses, playStatus)
				movieMap[movieID] = movie
			}
			continue
		}

		// Create a map of seen movie IDs for this user
		seenMovieIDs := make(map[string]bool)
		for _, seenMovie := range seenMovies {
			seenMovieIDs[seenMovie.ID] = true
		}

		// Update play status for each movie
		for movieID, movie := range movieMap {
			if seenMovieIDs[movieID] {
				// Movie is seen by this user, update play status
				playStatus := models.UserPlayStatus{
					UserID:   user.ID,
					UserName: user.Name,
					Played:   true,
					// Note: PlayCount would need to be fetched separately if needed
				}
				movie.UserPlayStatuses = append(movie.UserPlayStatuses, playStatus)
			} else {
				// Movie is NOT seen by this user, update play status
				playStatus := models.UserPlayStatus{
					UserID:   user.ID,
					UserName: user.Name,
					Played:   false,
				}
				movie.UserPlayStatuses = append(movie.UserPlayStatuses, playStatus)
			}
			movieMap[movieID] = movie
		}
	}

	// Convert map back to slice
	var moviesWithPlayStatus []models.Movie
	for _, movie := range movieMap {
		moviesWithPlayStatus = append(moviesWithPlayStatus, movie)
	}

	return moviesWithPlayStatus, nil
}

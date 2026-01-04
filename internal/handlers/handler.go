package handlers

import (
	"fmt"
	"jellyfin-duplicate/internal/jellyfin"
	"jellyfin-duplicate/internal/models"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	jellyfinClient *jellyfin.Client
}

func NewHandler(client *jellyfin.Client) *Handler {
	return &Handler{jellyfinClient: client}
}

func (h *Handler) GetDuplicatesPage(c *gin.Context) {
	log.Println("Handling request for duplicates page")
	duplicates, err := h.findDuplicates()
	if err != nil {
		log.Printf("Error finding duplicates: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Printf("Found %d duplicate pairs", len(duplicates))

	// Add play status discrepancy information to each duplicate
	for i, dup := range duplicates {
		discrepancies := h.GetPlayStatusDiscrepancies(dup.Movie1, dup.Movie2)
		if len(discrepancies) > 0 {
			duplicates[i].PlayStatusDiscrepancies = discrepancies
			duplicates[i].HasPlayStatusDiscrepancy = true
		}
	}

	// Separate duplicates and mismatches for better UI organization
	var potentialDuplicates []models.DuplicateResult
	var potentialMismatches []models.DuplicateResult

	for _, dup := range duplicates {
		if dup.IsDuplicate {
			potentialDuplicates = append(potentialDuplicates, dup)
		} else {
			potentialMismatches = append(potentialMismatches, dup)
		}
	}

	log.Printf("Rendering duplicates page with %d potential duplicates and %d potential mismatches",
		len(potentialDuplicates), len(potentialMismatches))

	c.HTML(http.StatusOK, "duplicates.html", gin.H{
		"duplicates":          duplicates,
		"potentialDuplicates": potentialDuplicates,
		"potentialMismatches": potentialMismatches,
	})
}

// GetPlayStatusForAllUsers fetches play status for all users for a duplicate pair
func (h *Handler) GetPlayStatusForAllUsers(dup models.DuplicateResult) (models.DuplicateResult, error) {
	// Get all users
	users, err := h.jellyfinClient.GetAllUsers()
	if err != nil {
		return dup, fmt.Errorf("failed to get users: %v", err)
	}

	// Fetch play status for each movie for all users
	for _, user := range users {
		// For movie 1
		status1, err := h.jellyfinClient.GetUserPlayStatus(dup.Movie1.ID, user.ID)
		if err != nil {
			log.Printf("Error getting play status for movie %s, user %s: %v", dup.Movie1.ID, user.ID, err)
			continue
		}

		// For movie 2
		status2, err := h.jellyfinClient.GetUserPlayStatus(dup.Movie2.ID, user.ID)
		if err != nil {
			log.Printf("Error getting play status for movie %s, user %s: %v", dup.Movie2.ID, user.ID, err)
			continue
		}

		// Add to movie's user play status
		dup.Movie1.UserPlayStatuses = append(dup.Movie1.UserPlayStatuses, status1)
		dup.Movie2.UserPlayStatuses = append(dup.Movie2.UserPlayStatuses, status2)
	}

	return dup, nil
}

func (h *Handler) GetPlayStatusDiscrepancies(movie1, movie2 models.Movie) []models.PlayStatusDiscrepancy {
	var discrepancies []models.PlayStatusDiscrepancy

	// Create maps for quick lookup
	movie1SeenBy := make(map[string]bool)
	movie2SeenBy := make(map[string]bool)

	for _, status := range movie1.UserPlayStatuses {
		if status.Played {
			movie1SeenBy[status.UserID] = true
		}
	}

	for _, status := range movie2.UserPlayStatuses {
		if status.Played {
			movie2SeenBy[status.UserID] = true
		}
	}

	// Check if movie1 is seen by users who haven't seen movie2
	for _, status := range movie1.UserPlayStatuses {
		if status.Played && !movie2SeenBy[status.UserID] {
			discrepancies = append(discrepancies, models.PlayStatusDiscrepancy{
				UserID:        status.UserID,
				UserName:      status.UserName,
				MovieToUpdate: movie2.ID,
				MovieName:     movie2.Name,
			})
		}
	}

	// Check if movie2 is seen by users who haven't seen movie1
	for _, status := range movie2.UserPlayStatuses {
		if status.Played && !movie1SeenBy[status.UserID] {
			discrepancies = append(discrepancies, models.PlayStatusDiscrepancy{
				UserID:        status.UserID,
				UserName:      status.UserName,
				MovieToUpdate: movie1.ID,
				MovieName:     movie1.Name,
			})
		}
	}

	return discrepancies
}

// HasIdenticalPlayStatus checks if two movies have identical play status for all users
func (h *Handler) HasIdenticalPlayStatus(movie1, movie2 models.Movie) bool {
	// If either movie has no play status data, they're not identical
	if len(movie1.UserPlayStatuses) == 0 || len(movie2.UserPlayStatuses) == 0 {
		return false
	}

	// Create maps for quick comparison
	movie1Status := make(map[string]bool) // userID -> played status
	movie2Status := make(map[string]bool) // userID -> played status

	// Build status maps for both movies
	for _, status := range movie1.UserPlayStatuses {
		movie1Status[status.UserID] = status.Played
	}

	for _, status := range movie2.UserPlayStatuses {
		movie2Status[status.UserID] = status.Played
	}

	// Check if they have the same users
	if len(movie1Status) != len(movie2Status) {
		return false
	}

	// Check if play status is identical for all users
	for userID, played1 := range movie1Status {
		played2, exists := movie2Status[userID]
		if !exists || played1 != played2 {
			return false
		}
	}

	return true
}

// GetMultiUserPlayStatus fetches play status for all users using the optimized approach
func (h *Handler) GetMultiUserPlayStatus() ([]models.Movie, error) {
	// Get all movies
	allMovies, err := h.jellyfinClient.GetAllMovies()
	if err != nil {
		return nil, fmt.Errorf("failed to get all movies: %v", err)
	}

	// Get all users
	users, err := h.jellyfinClient.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}

	// Fetch seen movies for all users in parallel
	userSeenMovies, err := h.jellyfinClient.GetSeenMoviesForAllUsers(users)
	if err != nil {
		return nil, fmt.Errorf("failed to get seen movies for all users: %v", err)
	}

	// Reconcile play status with all movies
	moviesWithPlayStatus, err := h.jellyfinClient.ReconcilePlayStatusWithAllMovies(allMovies, userSeenMovies, users)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile play status: %v", err)
	}

	return moviesWithPlayStatus, nil
}

func (h *Handler) GetDuplicatesJSON(c *gin.Context) {
	log.Println("Handling request for duplicates JSON")
	duplicates, err := h.findDuplicates()
	if err != nil {
		log.Printf("Error finding duplicates for JSON response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Printf("Returning %d duplicates in JSON format", len(duplicates))
	c.JSON(http.StatusOK, duplicates)
}

// DeleteMovie handles movie deletion requests
func (h *Handler) DeleteMovie(c *gin.Context) {
	movieID := c.Query("movieId")

	log.Printf("Received request to delete movie %s", movieID)

	// Validate required parameters
	if movieID == "" {
		log.Println("Invalid request: missing movieId parameter")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "movieId is a required parameter",
		})
		return
	}

	// Additional validation: check if movieID is valid format
	if len(movieID) < 32 || len(movieID) > 36 {
		log.Printf("Invalid movieId format: %s", movieID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid movieId format",
		})
		return
	}

	// Call Jellyfin API to delete the movie
	err := h.jellyfinClient.DeleteMovie(movieID)
	if err != nil {
		log.Printf("Failed to delete movie %s: %v", movieID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("failed to delete movie: %v", err).Error(),
		})
		return
	}

	log.Printf("Successfully deleted movie %s", movieID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Movie deleted successfully",
	})
}

// MarkMovieAsSeen marks a movie as seen for a specific user
func (h *Handler) MarkMovieAsSeen(c *gin.Context) {
	movieID := c.Query("movieId")
	userID := c.Query("userId")

	log.Printf("Received request to mark movie %s as seen for user %s", movieID, userID)

	// Validate required parameters
	if movieID == "" || userID == "" {
		log.Println("Invalid request: missing movieId or userId parameter")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "movieId and userId are required parameters",
		})
		return
	}

	// Additional validation: check if userID is valid format (UUID-like)
	if len(userID) < 32 || len(userID) > 36 {
		log.Printf("Invalid userId format: %s", userID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid userId format",
		})
		return
	}

	// Additional validation: check if movieID is valid format
	if len(movieID) < 32 || len(movieID) > 36 {
		log.Printf("Invalid movieId format: %s", movieID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid movieId format",
		})
		return
	}

	// Get movie and user names for better logging
	movieName := movieID // fallback to ID if name retrieval fails
	userName := userID   // fallback to ID if name retrieval fails

	if retrievedMovieName, err := h.jellyfinClient.GetMovieName(movieID); err == nil {
		movieName = retrievedMovieName
	}

	if retrievedUserName, err := h.jellyfinClient.GetUserName(userID); err == nil {
		userName = retrievedUserName
	}

	// Call Jellyfin API to mark movie as played
	err := h.jellyfinClient.MarkMovieAsPlayed(movieID, userID, movieName, userName)
	if err != nil {
		log.Printf("Failed to mark movie %s (%s) as played for user %s (%s): %v", movieName, movieID, userName, userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("failed to mark movie as played: %v", err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Movie marked as seen successfully",
	})
}

func (h *Handler) findDuplicates() ([]models.DuplicateResult, error) {
	log.Println("Starting duplicate detection process...")
	// Get all movies with multi-user play status from Jellyfin
	movies, err := h.GetMultiUserPlayStatus()
	if err != nil {
		return nil, err
	}

	log.Printf("Analyzing %d movies for duplicates", len(movies))

	var duplicates []models.DuplicateResult

	// Create a map to group movies by their Name and ProductionYear
	movieMap := make(map[string][]models.Movie)

	for _, movie := range movies {
		// Use Name-ProductionYear as the key
		// This handles cases where movies have the same name but different years
		key := fmt.Sprintf("%s-%d", movie.Name, movie.ProductionYear)

		movieMap[key] = append(movieMap[key], movie)
	}

	// Find duplicates by checking groups with more than one movie
	log.Printf("Found %d unique movie groups", len(movieMap))
	for _, group := range movieMap {
		if len(group) > 1 {
			// Compare all pairs in the group
			for i := 0; i < len(group); i++ {
				for j := i + 1; j < len(group); j++ {
					similarity := calculatePathSimilarity(group[i].Path, group[j].Path)
					isDuplicate := similarity >= 95

					// Check if movies have identical play status
					hasIdenticalPlayStatus := h.HasIdenticalPlayStatus(group[i], group[j])

					duplicates = append(duplicates, models.DuplicateResult{
						Movie1:                 group[i],
						Movie2:                 group[j],
						IsDuplicate:            isDuplicate,
						Similarity:             similarity,
						HasIdenticalPlayStatus: hasIdenticalPlayStatus,
					})
				}
			}
		}
	}

	log.Printf("Duplicate detection completed. Found %d duplicate pairs", len(duplicates))
	return duplicates, nil
}

// calculatePathSimilarity computes the similarity percentage between two paths
// using the Levenshtein distance algorithm implemented in pure Go
// Note: File extensions are excluded from the comparison
func calculatePathSimilarity(path1, path2 string) int {
	// Remove file extensions before comparison
	path1WithoutExt := removeFileExtension(path1)
	path2WithoutExt := removeFileExtension(path2)

	// Implement Levenshtein distance algorithm
	distance := levenshteinDistance(path1WithoutExt, path2WithoutExt)

	// Calculate maximum possible distance
	maxLen := len(path1WithoutExt)
	if len(path2WithoutExt) > maxLen {
		maxLen = len(path2WithoutExt)
	}

	if maxLen == 0 {
		return 100
	}

	// Calculate similarity percentage
	similarity := 100 - (distance * 100 / maxLen)
	return similarity
}

// removeFileExtension removes the file extension from a path
// Example: "/movies/movie.mkv" → "/movies/movie"
func removeFileExtension(path string) string {
	// Find the last dot in the path
	lastDotIndex := strings.LastIndex(path, ".")

	// If no dot found, or dot is at the start, return original path
	if lastDotIndex <= 0 {
		return path
	}

	// Check if the dot is part of a file extension
	// Look for common extension patterns
	lastSlashIndex := strings.LastIndex(path, "/")

	// If there's a slash after the last dot, it's not an extension
	if lastSlashIndex > lastDotIndex {
		return path
	}

	// Remove everything after the last dot (the extension)
	return path[:lastDotIndex]
}

// levenshteinDistance calculates the Levenshtein distance between two strings
// This is a pure Go implementation without external dependencies
func levenshteinDistance(s1, s2 string) int {
	// Convert strings to runes for proper Unicode handling
	r1 := []rune(s1)
	r2 := []rune(s2)

	len1 := len(r1)
	len2 := len(r2)

	// Create a matrix to store distances
	distances := make([][]int, len1+1)
	for i := range distances {
		distances[i] = make([]int, len2+1)
	}

	// Initialize the matrix
	for i := 0; i <= len1; i++ {
		distances[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		distances[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}

			distances[i][j] = min(
				distances[i-1][j]+1,      // deletion
				distances[i][j-1]+1,      // insertion
				distances[i-1][j-1]+cost, // substitution
			)
		}
	}

	return distances[len1][len2]
}

// min returns the minimum of multiple integers
func min(values ...int) int {
	if len(values) == 0 {
		return 0
	}

	minVal := values[0]
	for _, val := range values[1:] {
		if val < minVal {
			minVal = val
		}
	}
	return minVal
}

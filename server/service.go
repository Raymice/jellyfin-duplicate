package server

import (
	"fmt"
	jellyfinClients "jellyfin-duplicate/client/jellyfin/http"
	jellyfinModels "jellyfin-duplicate/client/jellyfin/models"
	"jellyfin-duplicate/utils"

	"github.com/sirupsen/logrus"
)

type ServerService struct {
	jellyfinClient *jellyfinClients.Client
}

func NewService(client *jellyfinClients.Client) *ServerService {
	return &ServerService{jellyfinClient: client}
}

// GetMultiUserPlayStatus fetches play status for all users using the optimized approach
func (s *ServerService) GetMultiUserPlayStatus() ([]jellyfinModels.Movie, error) {
	// Get all movies
	allMovies, err := s.jellyfinClient.GetAllMovies()
	if err != nil {
		return nil, fmt.Errorf("failed to get all movies: %v", err)
	}

	// Get all users
	users, err := s.jellyfinClient.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}

	// Fetch seen movies for all users in parallel
	userSeenMovies, err := s.jellyfinClient.GetSeenMoviesForAllUsers(users)
	if err != nil {
		return nil, fmt.Errorf("failed to get seen movies for all users: %v", err)
	}

	// Reconcile play status with all movies
	moviesWithPlayStatus, err := s.jellyfinClient.ReconcilePlayStatusWithAllMovies(allMovies, userSeenMovies, users)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile play status: %v", err)
	}

	return moviesWithPlayStatus, nil
}

func (s *ServerService) FindDuplicates() ([]jellyfinModels.DuplicateResult, error) {
	logrus.Info("Starting duplicate detection process...")
	// Get all movies with multi-user play status from Jellyfin
	movies, err := s.GetMultiUserPlayStatus()
	if err != nil {
		return nil, err
	}

	logrus.Infof("Analyzing %d movies for duplicates", len(movies))

	var duplicates []jellyfinModels.DuplicateResult

	// Create a map to group movies by their Name and ProductionYear
	movieMap := make(map[string][]jellyfinModels.Movie)

	for _, movie := range movies {
		// Use Name-ProductionYear as the key
		// This handles cases where movies have the same name but different years
		key := fmt.Sprintf("%s-%d", movie.Name, movie.ProductionYear)

		movieMap[key] = append(movieMap[key], movie)
	}

	// Find duplicates by checking groups with more than one movie
	logrus.Infof("Found %d unique movie groups", len(movieMap))
	for _, group := range movieMap {
		if len(group) > 1 {
			// Compare all pairs in the group
			for i := 0; i < len(group); i++ {
				for j := i + 1; j < len(group); j++ {
					similarity := utils.CalculatePathSimilarity(group[i].Path, group[j].Path)
					isDuplicate := similarity >= 95

					// Check if movies have identical play status
					hasIdenticalPlayStatus := s.HasIdenticalPlayStatus(group[i], group[j])

					duplicates = append(duplicates, jellyfinModels.DuplicateResult{
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

	logrus.Infof("Duplicate detection completed. Found %d duplicate pairs", len(duplicates))
	return duplicates, nil
}

// HasIdenticalPlayStatus checks if two movies have identical play status for all users
func (s *ServerService) HasIdenticalPlayStatus(movie1, movie2 jellyfinModels.Movie) bool {
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

// GetPlayStatusForAllUsers fetches play status for all users for a duplicate pair
func (s *ServerService) GetPlayStatusForAllUsers(dup jellyfinModels.DuplicateResult) (jellyfinModels.DuplicateResult, error) {
	// Get all users
	users, err := s.jellyfinClient.GetAllUsers()
	if err != nil {
		return dup, fmt.Errorf("failed to get users: %v", err)
	}

	// Fetch play status for each movie for all users
	for _, user := range users {
		// For movie 1
		status1, err := s.jellyfinClient.GetUserPlayStatus(dup.Movie1.ID, user.ID)
		if err != nil {
			logrus.Warnf("Error getting play status for movie %s, user %s: %v", dup.Movie1.ID, user.ID, err)
			continue
		}

		// For movie 2
		status2, err := s.jellyfinClient.GetUserPlayStatus(dup.Movie2.ID, user.ID)
		if err != nil {
			logrus.Warnf("Error getting play status for movie %s, user %s: %v", dup.Movie2.ID, user.ID, err)
			continue
		}

		// Add to movie's user play status
		dup.Movie1.UserPlayStatuses = append(dup.Movie1.UserPlayStatuses, status1)
		dup.Movie2.UserPlayStatuses = append(dup.Movie2.UserPlayStatuses, status2)
	}

	return dup, nil
}

func (s *ServerService) GetPlayStatusDiscrepancies(movie1, movie2 jellyfinModels.Movie) []jellyfinModels.PlayStatusDiscrepancy {
	var discrepancies []jellyfinModels.PlayStatusDiscrepancy

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
			discrepancies = append(discrepancies, jellyfinModels.PlayStatusDiscrepancy{
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
			discrepancies = append(discrepancies, jellyfinModels.PlayStatusDiscrepancy{
				UserID:        status.UserID,
				UserName:      status.UserName,
				MovieToUpdate: movie1.ID,
				MovieName:     movie1.Name,
			})
		}
	}

	return discrepancies
}

func (s *ServerService) DeleteMovie(movieID string) error {

	// Call Jellyfin API to delete the movie
	err := s.jellyfinClient.DeleteMovie(movieID)
	if err != nil {
		logrus.Errorf("Failed to delete movie %s: %v", movieID, err)
		return fmt.Errorf("failed to delete movie: %v", err)
	}

	return nil
}

func (s *ServerService) MarkMovieAsSeen(movieID, userID string) error {

	// Get movie and user names for better logging
	movieName := movieID // fallback to ID if name retrieval fails
	userName := userID   // fallback to ID if name retrieval fails

	if retrievedMovieName, err := s.jellyfinClient.GetMovieName(movieID); err == nil {
		movieName = retrievedMovieName
	}

	if retrievedUserName, err := s.jellyfinClient.GetUserName(userID); err == nil {
		userName = retrievedUserName
	}

	// Call Jellyfin API to mark movie as played
	err := s.jellyfinClient.MarkMovieAsPlayed(movieID, userID, movieName, userName)
	if err != nil {
		logrus.Errorf("Failed to mark movie %s (%s) as played for user %s (%s): %v", movieName, movieID, userName, userID, err)
		return fmt.Errorf("failed to mark movie as played: %v", err)
	}

	return nil
}

func IsUUIDFormtatted(id string) bool {
	if len(id) < 32 || len(id) > 36 {
		return false
	} else {
		return true
	}
}

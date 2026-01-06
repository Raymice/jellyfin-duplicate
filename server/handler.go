package server

import (
	"fmt"
	jellyfinClients "jellyfin-duplicate/client/jellyfin/http"
	jellyfinModels "jellyfin-duplicate/client/jellyfin/models"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	serverService *ServerService
}

func NewHandler(client *jellyfinClients.Client) *Handler {
	serverService := NewService(client)
	return &Handler{serverService: serverService}
}

// GET /
func (h *Handler) GetHomePage(ctx *gin.Context) {
	logrus.Info("Handling request for home page")
	ctx.HTML(http.StatusOK, "home.html", gin.H{})
}

// GET /analysis
func (h *Handler) GetDuplicatesPage(ctx *gin.Context) {
	logrus.Info("Handling request for duplicates page")
	duplicates, err := h.serverService.FindDuplicates()
	if err != nil {
		logrus.Errorf("Error finding duplicates: %v", err)
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("Found %d duplicate pairs", len(duplicates))

	// Add play status discrepancy information to each duplicate
	for i, dup := range duplicates {
		discrepancies := h.serverService.GetPlayStatusDiscrepancies(dup.Movie1, dup.Movie2)
		if len(discrepancies) > 0 {
			duplicates[i].PlayStatusDiscrepancies = discrepancies
			duplicates[i].HasPlayStatusDiscrepancy = true
		}
	}

	// Separate duplicates and mismatches for better UI organization
	var potentialDuplicates []jellyfinModels.DuplicateResult
	var potentialMismatches []jellyfinModels.DuplicateResult

	for _, dup := range duplicates {
		if dup.IsDuplicate {
			potentialDuplicates = append(potentialDuplicates, dup)
		} else {
			potentialMismatches = append(potentialMismatches, dup)
		}
	}

	logrus.Infof("Rendering duplicates page with %d potential duplicates and %d potential mismatches",
		len(potentialDuplicates), len(potentialMismatches))

	ctx.HTML(http.StatusOK, "duplicates.html", gin.H{
		"duplicates":          duplicates,
		"potentialDuplicates": potentialDuplicates,
		"potentialMismatches": potentialMismatches,
	})
}

// GET /api/duplicates
func (h *Handler) GetDuplicatesJSON(ctx *gin.Context) {
	logrus.Info("Handling request for duplicates JSON")
	duplicates, err := h.serverService.FindDuplicates()
	if err != nil {
		logrus.Errorf("Error finding duplicates for JSON response: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("Returning %d duplicates in JSON format", len(duplicates))
	ctx.JSON(http.StatusOK, duplicates)
}

// GET /api/delete-movie
// DeleteMovie handles movie deletion requests
func (h *Handler) DeleteMovie(ctx *gin.Context) {
	movieID := ctx.Query("movieId")

	logrus.Infof("Received request to delete movie %s", movieID)

	// Validate required parameters
	if lo.IsEmpty(movieID) {
		logrus.Warn("Invalid request: missing movieId parameter")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "movieId is a required parameter",
		})
		return
	}

	// Additional validation: check if movieID is valid format
	if !IsUUIDFormtatted(movieID) {
		logrus.Warnf("Invalid movieId format: %s", movieID)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid movieId format",
		})
		return
	}

	err := h.serverService.DeleteMovie(movieID)
	if err != nil {
		logrus.Errorf("Error deleting movie %s: %v", movieID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.Infof("Successfully deleted movie %s", movieID)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Movie deleted successfully",
	})
}

// GET /api/mark-as-seen
// MarkMovieAsSeen marks a movie as seen for a specific user
func (h *Handler) MarkMovieAsSeen(ctx *gin.Context) {
	movieID := ctx.Query("movieId")
	userID := ctx.Query("userId")

	logrus.Infof("Received request to mark movie %s as seen for user %s", movieID, userID)

	// Validate required parameters
	if lo.IsEmpty(movieID) || lo.IsEmpty(userID) {
		logrus.Warn("Invalid request: missing movieId or userId parameter")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "movieId and userId are required parameters",
		})
		return
	}

	// Additional validation: check if userID is valid format (UUID-like)
	if !IsUUIDFormtatted(userID) {
		logrus.Warnf("Invalid userId format: %s", userID)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid userId format",
		})
		return
	}

	// Additional validation: check if movieID is valid format
	if !IsUUIDFormtatted(movieID) {
		logrus.Warnf("Invalid movieId format: %s", movieID)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid movieId format",
		})
		return
	}

	err := h.serverService.MarkMovieAsSeen(movieID, userID)

	if err != nil {
		logrus.Errorf("Failed to mark movie %s as seen for user %s: %v", movieID, userID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("failed to mark movie as seen: %v", err).Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Movie marked as seen successfully",
	})
}

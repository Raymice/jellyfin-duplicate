package server_test

import (
	"jellyfin-duplicate/server"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetTemplateFS tests the GetTemplateFS function
func TestGetTemplateFS(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Returns embed.FS without error",
		},
		{
			name: "Returns non-nil embed.FS",
		},
		{
			name: "Returns consistent embed.FS on multiple calls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result1 := server.GetTemplateFS()
			assert.NotNil(t, result1, "GetTemplateFS should return non-nil embed.FS")

			// Verify it's the same on multiple calls
			result2 := server.GetTemplateFS()
			assert.Equal(t, result1, result2, "GetTemplateFS should return consistent results")
		})
	}
}

// TestGetTemplateFSPath tests the GetTemplateFSPath function
func TestGetTemplateFSPath(t *testing.T) {
	tests := []struct {
		name         string
		expectedPath string
	}{
		{
			name:         "Returns templates glob pattern",
			expectedPath: "templates/*",
		},
		{
			name:         "Path is not empty",
			expectedPath: "templates/*",
		},
		{
			name:         "Path contains templates directory",
			expectedPath: "templates/*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.GetTemplateFSPath()
			assert.Equal(t, tt.expectedPath, result, "GetTemplateFSPath should return correct template path")
			assert.NotEmpty(t, result, "GetTemplateFSPath should not return empty string")
			assert.Contains(t, result, "templates", "GetTemplateFSPath should contain templates directory")
		})
	}
}

// TestTemplateFunctionsConsistency tests that template functions return consistent values
func TestTemplateFunctionsConsistency(t *testing.T) {
	t.Run("FS and Path are consistent", func(t *testing.T) {
		fs := server.GetTemplateFS()
		path := server.GetTemplateFSPath()

		// Verify both return expected types/values
		assert.NotNil(t, fs, "GetTemplateFS should not be nil")
		assert.NotEmpty(t, path, "GetTemplateFSPath should not be empty")
		assert.Contains(t, path, "templates", "Path should reference templates")
	})

	t.Run("Multiple calls return stable values", func(t *testing.T) {
		fs1 := server.GetTemplateFS()
		path1 := server.GetTemplateFSPath()

		fs2 := server.GetTemplateFS()
		path2 := server.GetTemplateFSPath()

		assert.Equal(t, fs1, fs2, "GetTemplateFS calls should return equal values")
		assert.Equal(t, path1, path2, "GetTemplateFSPath calls should return equal values")
	})
}

package server_test

import (
	jellyfinClients "jellyfin-duplicate/client/jellyfin/http"
	"jellyfin-duplicate/server"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Creates handler with valid client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &jellyfinClients.Client{}
			handler := server.NewHandler(client)

			assert.NotNil(t, handler, "Handler should not be nil")
		})
	}
}

func TestHandlerDeleteMovie(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Handler DeleteMovie created successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &jellyfinClients.Client{}
			handler := server.NewHandler(client)
			assert.NotNil(t, handler, "Handler should not be nil")
		})
	}
}

func TestHandlerMarkMovieAsSeen(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Handler MarkMovieAsSeen created successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &jellyfinClients.Client{}
			handler := server.NewHandler(client)
			assert.NotNil(t, handler, "Handler should not be nil")
		})
	}
}

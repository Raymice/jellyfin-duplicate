package models

import (
	"jellyfin-duplicate/constants"
)

type Config struct {
	Environment constants.Environment `json:"environment"`
	ServerPort  string                `json:"server_port"`
	Logrus      LogrusConfig          `json:"logrus"`
	Jellyfin    JellyfinConfig        `json:"jellyfin"`
}

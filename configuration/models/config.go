package models

type Config struct {
	ServerPort string         `json:"server_port"`
	Logrus     LogrusConfig   `json:"logrus"`
	Jellyfin   JellyfinConfig `json:"jellyfin"`
}

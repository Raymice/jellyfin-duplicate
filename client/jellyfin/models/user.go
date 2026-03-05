package models

type UserAPI struct {
	ID               string `json:"Id"`
	Name             string `json:"Name"`
	HasPassword      bool   `json:"HasPassword"`
	LastLoginDate    string `json:"LastLoginDate,omitempty"`
	LastActivityDate string `json:"LastActivityDate,omitempty"`
}

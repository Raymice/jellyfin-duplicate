package models

type Movie struct {
	ID             string `json:"Id"`
	Name           string `json:"Name"`
	Path           string `json:"Path"`
	ProductionYear int    `json:"ProductionYear"`
	PlayStatus     UserPlayStatus `json:"PlayStatus"`
	UserData       struct {
		Played         bool   `json:"Played"`
		PlaybackPositionTicks int64 `json:"PlaybackPositionTicks"`
		PlayCount      int    `json:"PlayCount"`
		LastPlayedDate string `json:"LastPlayedDate"`
	} `json:"UserData"`
	ProviderIds    struct {
		Tmdb string `json:"Tmdb"`
		Imdb string `json:"Imdb"`
	} `json:"ProviderIds"`
	UserPlayStatuses []UserPlayStatus `json:"UserPlayStatuses"`
}

type UserPlayStatus struct {
	UserID    string `json:"UserId"`
	UserName  string `json:"UserName"`
	Played    bool   `json:"Played"`
	PlayCount int    `json:"PlayCount"`
}

// User model for multi-user support
type User struct {
	ID          string `json:"Id"`
	Name        string `json:"Name"`
	HasPassword bool   `json:"HasPassword"`
	LastLoginDate string `json:"LastLoginDate,omitempty"`
	LastActivityDate string `json:"LastActivityDate,omitempty"`
}

// Extended Movie model with play status
type MovieWithPlayStatus struct {
	Movie
	PlayStatus     UserPlayStatus `json:"PlayStatus"`
	UserPlayStatuses []UserPlayStatus `json:"UserPlayStatuses"`
}

// PlayStatusDiscrepancy represents a discrepancy in play status between duplicate movies
type PlayStatusDiscrepancy struct {
	UserID        string `json:"user_id"`
	UserName      string `json:"user_name"`
	MovieToUpdate string `json:"movie_to_update"`
	MovieName     string `json:"movie_name"`
}

type DuplicateResult struct {
	Movie1                  Movie                  `json:"movie1"`
	Movie2                  Movie                  `json:"movie2"`
	IsDuplicate             bool                   `json:"is_duplicate"`
	Similarity              int                    `json:"similarity"`
	HasPlayStatusDiscrepancy bool                  `json:"has_play_status_discrepancy"`
	HasIdenticalPlayStatus  bool                  `json:"has_identical_play_status"`
	PlayStatusDiscrepancies []PlayStatusDiscrepancy `json:"play_status_discrepancies,omitempty"`
}
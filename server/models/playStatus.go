package models

type UserPlayStatusDTO struct {
	UserID    string `json:"UserId"`
	UserName  string `json:"UserName"`
	Played    bool   `json:"Played"`
	PlayCount int    `json:"PlayCount"`
}

// PlayStatusDiscrepancy represents a discrepancy in play status between duplicate movies
type PlayStatusDiscrepancyDTO struct {
	UserID        string `json:"user_id"`
	UserName      string `json:"user_name"`
	MovieToUpdate string `json:"movie_to_update"`
	MovieName     string `json:"movie_name"`
}

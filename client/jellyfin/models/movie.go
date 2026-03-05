package models

type MovieXtLightAPI struct {
	ID             string `json:"Id"`
	Name           string `json:"Name"`
	ProductionYear int    `json:"ProductionYear"`
}

type MovieLightAPI struct {
	MovieXtLightAPI
	Path string `json:"Path"`
}

type UserDataAPI struct {
	Played                bool   `json:"Played"`
	PlaybackPositionTicks int64  `json:"PlaybackPositionTicks"`
	PlayCount             int    `json:"PlayCount"`
	LastPlayedDate        string `json:"LastPlayedDate"`
}

type MediaStreamsAPI struct {
	DisplayTitle string `json:"DisplayTitle"`
	Type         string `json:"Type"`
}

type MovieAPI struct {
	MovieXtLightAPI
	UserData     UserDataAPI       `json:"UserData"`
	MediaStreams []MediaStreamsAPI `json:"MediaStreams"`
}

type MovieLightExtendedAPI struct {
	MovieLightAPI
	MediaStreams []MediaStreamsAPI `json:"MediaStreams"`
}

// type Movie struct {
// 	MovieLightStatus
// 	// TODO review PlayStatus, split API models and UI models if needed
// 	PlayStatus UserPlayStatus `json:"PlayStatus"`
// 	UserData   struct {
// 		Played                bool   `json:"Played"`
// 		PlaybackPositionTicks int64  `json:"PlaybackPositionTicks"`
// 		PlayCount             int    `json:"PlayCount"`
// 		LastPlayedDate        string `json:"LastPlayedDate"`
// 	} `json:"UserData"`
// 	MediaStreams []MediaStreamsAPI `json:"MediaStreams"`
// }

type MovieDTO struct {
	MovieLightExtendedAPI
	UserPlayStatuses []UserPlayStatus `json:"UserPlayStatuses"`
}

type MovieLightStatus struct {
	MovieLightAPI
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
	ID               string `json:"Id"`
	Name             string `json:"Name"`
	HasPassword      bool   `json:"HasPassword"`
	LastLoginDate    string `json:"LastLoginDate,omitempty"`
	LastActivityDate string `json:"LastActivityDate,omitempty"`
}

// Extended Movie model with play status
// type MovieWithPlayStatus struct {
// 	Movie
// 	PlayStatus       UserPlayStatus   `json:"PlayStatus"`
// 	UserPlayStatuses []UserPlayStatus `json:"UserPlayStatuses"`
// }

// PlayStatusDiscrepancy represents a discrepancy in play status between duplicate movies
type PlayStatusDiscrepancy struct {
	UserID        string `json:"user_id"`
	UserName      string `json:"user_name"`
	MovieToUpdate string `json:"movie_to_update"`
	MovieName     string `json:"movie_name"`
}

type DuplicateResult struct {
	Movie1                   MovieDTO                `json:"movie1"`
	Movie2                   MovieDTO                `json:"movie2"`
	IsDuplicate              bool                    `json:"is_duplicate"`
	Similarity               int                     `json:"similarity"`
	HasPlayStatusDiscrepancy bool                    `json:"has_play_status_discrepancy"`
	HasIdenticalPlayStatus   bool                    `json:"has_identical_play_status"`
	PlayStatusDiscrepancies  []PlayStatusDiscrepancy `json:"play_status_discrepancies,omitempty"`
}

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

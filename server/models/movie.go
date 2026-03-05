package models

import apiModels "jellyfin-duplicate/client/jellyfin/models"

type MovieDTO struct {
	apiModels.MovieLightExtendedAPI
	UserPlayStatuses []UserPlayStatusDTO `json:"UserPlayStatuses"`
}

type MovieLightStatusDTO struct {
	apiModels.MovieLightAPI
	UserPlayStatuses []UserPlayStatusDTO `json:"UserPlayStatuses"`
}

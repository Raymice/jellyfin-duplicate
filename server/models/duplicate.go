package models

type DuplicateResultDTO struct {
	Movie1                   MovieDTO                   `json:"movie1"`
	Movie2                   MovieDTO                   `json:"movie2"`
	IsDuplicate              bool                       `json:"is_duplicate"`
	Similarity               int                        `json:"similarity"`
	HasPlayStatusDiscrepancy bool                       `json:"has_play_status_discrepancy"`
	HasIdenticalPlayStatus   bool                       `json:"has_identical_play_status"`
	PlayStatusDiscrepancies  []PlayStatusDiscrepancyDTO `json:"play_status_discrepancies,omitempty"`
}

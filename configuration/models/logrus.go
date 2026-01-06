package models

type LogrusConfig struct {
	Level         string `json:"level"`
	Format        string `json:"format"`
	DisableColors bool   `json:"disable_colors"`
	ReportCaller  bool   `json:"report_caller"`
}

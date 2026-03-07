package server

import "embed"

//go:embed  templates/*
var templatesFS embed.FS

func GetTemplateFS() embed.FS {
	return templatesFS
}

func GetTemplateFSPath() string {
	return "templates/*"
}

package web

import (
	"embed"
)

//go:embed templates/index.html
var TemplatesFS embed.FS

//go:embed static/*
var StaticFS embed.FS

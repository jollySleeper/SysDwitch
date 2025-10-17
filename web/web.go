package web

import (
	"embed"
)

//go:embed templates/index.html
var TemplatesFS embed.FS

//go:embed static/css static/js
var StaticFS embed.FS

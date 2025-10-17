package web

import (
	"embed"
)

//go:embed templates/index.html
var TemplatesFS embed.FS

// TODO: Add static files embedding when static assets are available
// var StaticFS embed.FS

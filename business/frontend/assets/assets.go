package assets

import (
	"embed"
	"net/http"
)

//go:embed static
var static embed.FS

var StaticHandler = http.FileServer(http.FS(static))

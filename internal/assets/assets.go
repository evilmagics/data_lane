package assets

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// GetEmbeddedAssets returns the filesystem for the embedded frontend
func GetEmbeddedAssets() fs.FS {
	// Strip the "dist" prefix so that "dist/index.html" becomes "index.html"
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}

package templates

import (
	"embed"
	"io/fs"
)

//go:embed *.tpl
var fsys embed.FS

func FS() fs.FS {
	return fsys
}

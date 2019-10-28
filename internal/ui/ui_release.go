// +build release

package ui

import (
	_ "rischmann.fr/apero/statik"

	"github.com/rakyll/statik/fs"
)

// GetFile returns the content of the file at path
func GetFile(path string) ([]byte, error) {
	if path[0] != '/' {
		panic("path should start with /")
	}

	statikFS, err := fs.New()
	if err != nil {
		return err
	}

	return fs.ReadFile(statikFS, path)
}

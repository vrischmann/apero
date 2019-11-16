// +build release

package ui

import (
	"fmt"
	"html/template"

	_ "rischmann.fr/apero/statik"

	"github.com/rakyll/statik/fs"
)

// GetFile returns the content of the file at path
func GetFile(path string) ([]byte, error) {
	if path[0] != '/' {
		panic(fmt.Errorf("path %q should start with /", path))
	}

	statikFS, err := fs.New()
	if err != nil {
		return nil, err
	}

	return fs.ReadFile(statikFS, path)
}

// ParseTemplate parses the content of the files in `paths` into a template.
// If there's any error the function panics.
func ParseTemplate(funcs template.FuncMap, paths ...string) *template.Template {
	tmpl := template.New("root").Funcs(funcs)

	for _, path := range paths {
		data := MustGetFile(path)
		tmpl = template.Must(tmpl.Parse(string(data)))
	}

	return tmpl
}

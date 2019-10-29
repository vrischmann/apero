// +build !release

package ui

import (
	"fmt"
	"html/template"
	"io/ioutil"
)

// GetFile returns the content of the file at path
func GetFile(path string) ([]byte, error) {
	if path[0] != '/' {
		panic(fmt.Errorf("path %q should start with /", path))
	}

	path = "./ui" + path

	return ioutil.ReadFile(path)
}

// ParseTemplate parses the content of the files in `paths` into a template.
// If there's any error the function panics.
func ParseTemplate(funcs template.FuncMap, paths ...string) *template.Template {
	for i := range paths {
		if paths[i][0] != '/' {
			panic(fmt.Errorf("path %q should start with /", paths[i]))
		}
		paths[i] = "./ui" + paths[i]
	}

	return template.Must(template.New("root").Funcs(funcs).ParseFiles(paths...))
}

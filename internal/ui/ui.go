package ui

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"time"
)

// MustGetFile returns the content of the file at `path`.
// If there's any error it panics.
func MustGetFile(path string) []byte {
	data, err := GetFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

// ServeFile returns a http.HandlerFunc which serves the content of the file at `path`.
func ServeFile(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		data := MustGetFile(path)
		http.ServeContent(w, req, path, time.Now(), bytes.NewReader(data))
	}
}

// ParseTemplate parses the content of the file at `path` as a html.Template.
// If there's any error it panics.
func ParseTemplate(path string) *template.Template {
	data := MustGetFile(path)
	tmpl, err := template.New("root").Parse(string(data))
	if err != nil {
		log.Fatal(err)
	}
	return tmpl
}

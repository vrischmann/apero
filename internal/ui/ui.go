package ui

import (
	"bytes"
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

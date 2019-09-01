package main

import (
	"net/http"
	"path"
	"strings"
)

// shiftPath splits the path into the first component and the rest of
// the path.
// The returned head will never have a slash in it, if the path has no tail
// head will be empty.
// The tail will never have a trailing slash.
func shiftPath(p string) (head string, tail string) {
	p = path.Clean("/" + p)

	pos := strings.Index(p[1:], "/")
	if pos == -1 {
		return p[1:], "/"
	}

	p = p[1:]

	return p[:pos], p[pos:]
}

type server struct {
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	head, _ := shiftPath(req.URL.Path)
	switch head {
	case "copy":
		s.handleCopy(w, req)
	case "move":
		s.handleMove(w, req)
	case "paste":
		s.handlePaste(w, req)
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (s *server) handleCopy(w http.ResponseWriter, req *http.Request) {
}

func (s *server) handleMove(w http.ResponseWriter, req *http.Request) {
}

func (s *server) handlePaste(w http.ResponseWriter, req *http.Request) {
}

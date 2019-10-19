package main

import (
	"net/http"

	"github.com/oklog/ulid/v2"
)

func isEmptyULID(id ulid.ULID) bool {
	var emptyID ulid.ULID
	return id == emptyID
}

func responseStatusCode(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	w.Write([]byte(http.StatusText(code)))
}

func responseString(w http.ResponseWriter, s string, code int) {
	w.WriteHeader(code)
	w.Write([]byte(s))
}

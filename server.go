package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"path"
	"strings"

	"rischmann.fr/apero/internal"
)

type serverConfig struct {
	ListenAddr string
	Key        internal.SecretBoxKey
}

func (c serverConfig) Validate() error {
	if _, _, err := net.SplitHostPort(c.ListenAddr); err != nil {
		return err
	}
	return nil
}

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
	store *store
}

func newServer(conf serverConfig) *server {
	return &server{}
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
	case "register":
		s.handleRegister(w, req)
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (s *server) handleCopy(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		responseStatusCode(w, http.StatusMethodNotAllowed)
		return
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		responseStatusCode(w, http.StatusInternalServerError)
		return
	}

	//

	var payload copyRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("unable to unmarshal copy request payload. err=%v", err)
		responseString(w, "invalid copy request", http.StatusBadRequest)
		return
	}
	if err := payload.Validate(); err != nil {
		log.Printf("copy request payload invalid. err=%v", err)
		responseString(w, "invalid copy request", http.StatusBadRequest)
		return
	}

	//

	pk, err := s.store.LookupPublicKey(payload.DeviceID)
	switch {
	case err == errDeviceIDNotFound:
		log.Printf("device id is not registered. err=%v", err)
		responseStatusCode(w, http.StatusNotFound)
	case err != nil:
		log.Printf("unable to lookup public key. err=%v", err)
		responseStatusCode(w, http.StatusInternalServerError)
	}

	validSignature := internal.VerifySignature(pk, payload.Content, payload.Signature)
	if !validSignature {
		log.Printf("invalid signature for device %s", payload.DeviceID)
		responseString(w, "invalid signature", http.StatusBadRequest)
		return
	}
}

func (s *server) handleMove(w http.ResponseWriter, req *http.Request) {
}

func (s *server) handlePaste(w http.ResponseWriter, req *http.Request) {
}

func (s *server) handleRegister(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		responseStatusCode(w, http.StatusMethodNotAllowed)
		return
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		responseStatusCode(w, http.StatusInternalServerError)
		return
	}

	var payload registerRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("unable to unmarshal copy request payload. err=%v", err)
		responseString(w, "invalid copy request", http.StatusBadRequest)
		return
	}
	if err := payload.Validate(); err != nil {
		log.Printf("copy request payload invalid. err=%v", err)
		responseString(w, "invalid copy request", http.StatusBadRequest)
		return
	}

	//

	log.Printf("payload: %+v", payload)
}

func responseStatusCode(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	w.Write([]byte(http.StatusText(code)))
}

func responseString(w http.ResponseWriter, s string, code int) {
	w.WriteHeader(code)
	w.Write([]byte(s))
}

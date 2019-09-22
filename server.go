package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/oklog/ulid/v2"
	"github.com/vrischmann/hutil/v2"
)

type serverConfig struct {
	ListenAddr    string
	PSKey         secretBoxKey
	SignPublicKey publicKey
}

func (c serverConfig) Validate() error {
	if _, _, err := net.SplitHostPort(c.ListenAddr); err != nil {
		return err
	}
	if !c.PSKey.IsValid() {
		return fmt.Errorf("ps key is invalid")
	}
	if !c.SignPublicKey.IsValid() {
		return fmt.Errorf("sign public key is invalid")
	}
	return nil
}

type server struct {
	conf serverConfig
	st   store
}

func newServer(conf serverConfig, st store) *server {
	return &server{
		conf: conf,
		st:   st,
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	head, _ := hutil.ShiftPath(req.URL.Path)
	switch head {
	case "copy":
		s.handleCopy(w, req)
	case "move", "paste":
		s.handleMoveOrPaste(w, req, head)
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
	defer req.Body.Close()

	data, ok := secretBoxOpen(data, s.conf.PSKey)
	if !ok {
		log.Printf("unable to open box")
		responseStatusCode(w, http.StatusBadRequest)
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

	validSignature := verify(s.conf.SignPublicKey, payload.Content, payload.Signature)
	if !validSignature {
		log.Printf("invalid signature")
		responseString(w, "invalid signature", http.StatusBadRequest)
		return
	}

	// TODO(vincent): size limits and stuff

	id, err := s.st.Add(payload.Content)
	if err != nil {
		log.Printf("unable to store payload. err=%v", err)
		responseString(w, "internal server error", http.StatusInternalServerError)
		return
	}

	respData := secretBoxSeal(id[:], s.conf.PSKey)

	w.WriteHeader(http.StatusAccepted)
	w.Write(respData)
}

func (s *server) handleMoveOrPaste(w http.ResponseWriter, req *http.Request, action string) {
	if req.Method != http.MethodPost {
		responseStatusCode(w, http.StatusMethodNotAllowed)
		return
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		responseStatusCode(w, http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	data, ok := secretBoxOpen(data, s.conf.PSKey)
	if !ok {
		log.Printf("unable to open box")
		responseStatusCode(w, http.StatusBadRequest)
		return
	}

	//

	var payload moveOrPasteRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("unable to unmarshal move request payload. err=%v", err)
		responseString(w, "invalid move request", http.StatusBadRequest)
		return
	}
	if err := payload.Validate(); err != nil {
		log.Printf("move request payload invalid. err=%v", err)
		responseString(w, "invalid move request", http.StatusBadRequest)
		return
	}

	//

	// NOTE(vincent): since there's no content in a move request we sign the single byte M
	validSignature := verify(s.conf.SignPublicKey, payload.ID[:], payload.Signature)
	if !validSignature {
		log.Printf("invalid signature")
		responseString(w, "invalid signature", http.StatusBadRequest)
		return
	}

	//

	var content []byte

	switch {
	case action == "move" && isEmptyULID(payload.ID):
		content, err = s.st.RemoveFirst()

	case action == "move" && !isEmptyULID(payload.ID):
		content, err = s.st.Remove(payload.ID)

	case action == "paste" && isEmptyULID(payload.ID):
		content, err = s.st.CopyFirst()

	case action == "paste" && !isEmptyULID(payload.ID):
		content, err = s.st.Copy(payload.ID)
	}

	if err != nil {
		log.Printf("unable to retrieve entry. err=%v", err)
		responseString(w, "internal server error", http.StatusInternalServerError)
		return
	}

	//

	respData := secretBoxSeal(content, s.conf.PSKey)

	w.WriteHeader(http.StatusAccepted)
	w.Write(respData)
}

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

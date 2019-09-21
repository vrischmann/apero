package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

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
}

func newServer(conf serverConfig) *server {
	return &server{
		conf: conf,
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	head, _ := hutil.ShiftPath(req.URL.Path)
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

	// TODO(vincent): store this shit

	// TODO(vincent): temporary for testing
	respData := secretBoxSeal([]byte("OK"), s.conf.PSKey)

	w.WriteHeader(http.StatusAccepted)
	w.Write(respData)
}

func (s *server) handleMove(w http.ResponseWriter, req *http.Request) {
}

func (s *server) handlePaste(w http.ResponseWriter, req *http.Request) {
}

func responseStatusCode(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	w.Write([]byte(http.StatusText(code)))
}

func responseString(w http.ResponseWriter, s string, code int) {
	w.WriteHeader(code)
	w.Write([]byte(s))
}

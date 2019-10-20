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

type apiHandler struct {
	conf serverConfig
	st   store
}

func newAPIHandler(conf serverConfig, st store) *apiHandler {
	return &apiHandler{
		conf: conf,
		st:   st,
	}
}

func (s *apiHandler) handle(w http.ResponseWriter, req *http.Request, path string) {
	head, tail := hutil.ShiftPath(path)
	if head != "v1" {
		http.Error(w, fmt.Sprintf("%q is not a valid version", head), http.StatusBadRequest)
		return
	}

	head, _ = hutil.ShiftPath(tail)

	switch head {
	case "copy":
		s.handleCopy(w, req)
	case "move":
		s.handleMove(w, req)
	case "paste":
		s.handlePaste(w, req)
	case "list":
		s.handleList(w, req)
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (s *apiHandler) handleCopy(w http.ResponseWriter, req *http.Request) {
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
		log.Printf("unable to unmarshal copy request payload. err: %v", err)
		responseString(w, "invalid copy request", http.StatusBadRequest)
		return
	}
	if err := payload.Validate(); err != nil {
		log.Printf("copy request payload invalid. err: %v", err)
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
		log.Printf("unable to store payload. err: %v", err)
		responseString(w, "internal server error", http.StatusInternalServerError)
		return
	}

	respData := secretBoxSeal(id[:], s.conf.PSKey)

	w.WriteHeader(http.StatusAccepted)
	w.Write(respData)
}

func (s *apiHandler) handleMove(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
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

	var payload moveRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("unable to unmarshal move request payload. err: %v", err)
		responseString(w, "invalid move request", http.StatusBadRequest)
		return
	}
	if err := payload.Validate(); err != nil {
		log.Printf("move request payload invalid. err: %v", err)
		responseString(w, "invalid move request", http.StatusBadRequest)
		return
	}

	//

	validSignature := verify(s.conf.SignPublicKey, payload.ID[:], payload.Signature)
	if !validSignature {
		log.Printf("invalid signature")
		responseString(w, "invalid signature", http.StatusBadRequest)
		return
	}

	//

	var content []byte

	if isEmptyULID(payload.ID) {
		content, err = s.st.RemoveFirst()
	} else {
		content, err = s.st.Remove(payload.ID)
	}

	if err != nil {
		log.Printf("unable to retrieve entry. err: %v", err)
		responseString(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(content) == 0 {
		responseStatusCode(w, http.StatusNotFound)
		return
	}

	//

	respData := secretBoxSeal(content, s.conf.PSKey)

	w.WriteHeader(http.StatusOK)
	w.Write(respData)
}

func (s *apiHandler) handlePaste(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
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

	var payload pasteRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("unable to unmarshal move request payload. err: %v", err)
		responseString(w, "invalid move request", http.StatusBadRequest)
		return
	}
	if err := payload.Validate(); err != nil {
		log.Printf("move request payload invalid. err: %v", err)
		responseString(w, "invalid move request", http.StatusBadRequest)
		return
	}

	//

	validSignature := verify(s.conf.SignPublicKey, payload.ID[:], payload.Signature)
	if !validSignature {
		log.Printf("invalid signature")
		responseString(w, "invalid signature", http.StatusBadRequest)
		return
	}

	//

	var content []byte

	if isEmptyULID(payload.ID) {
		content, err = s.st.CopyFirst()
	} else {
		content, err = s.st.Copy(payload.ID)
	}

	if err != nil {
		log.Printf("unable to retrieve entry. err: %v", err)
		responseString(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(content) == 0 {
		responseStatusCode(w, http.StatusNotFound)
		return
	}

	//

	respData := secretBoxSeal(content, s.conf.PSKey)

	w.WriteHeader(http.StatusOK)
	w.Write(respData)
}

func (s *apiHandler) handleList(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		responseStatusCode(w, http.StatusMethodNotAllowed)
		return
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		responseStatusCode(w, http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	if len(data) == 0 {
		log.Printf("no data in list request")
		responseStatusCode(w, http.StatusBadRequest)
		return
	}

	data, ok := secretBoxOpen(data, s.conf.PSKey)
	if !ok {
		log.Printf("unable to open box")
		responseStatusCode(w, http.StatusBadRequest)
		return
	}

	//

	var payload listRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("unable to unmarshal list request payload. err: %v", err)
		responseString(w, "invalid list request", http.StatusBadRequest)
		return
	}
	if err := payload.Validate(); err != nil {
		log.Printf("list request payload invalid. err: %v", err)
		responseString(w, "invalid list request", http.StatusBadRequest)
		return
	}

	//

	// NOTE(vincent): since there's no content in a move request we sign the single byte L
	// This might change in the future when we expand the protocol
	validSignature := verify(s.conf.SignPublicKey, []byte("L"), payload.Signature)
	if !validSignature {
		log.Printf("invalid signature")
		responseString(w, "invalid signature", http.StatusBadRequest)
		return
	}

	entries, err := s.st.ListAll()
	if err != nil {
		log.Printf("unable to list all entries. err: %v", err)
		responseString(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var resp listResponse
	resp.Entries = entries

	content, err := json.Marshal(resp)
	if err != nil {
		log.Printf("unable to marshal list response. err: %v", err)
		responseString(w, "internal server error", http.StatusInternalServerError)
		return
	}

	//

	respData := secretBoxSeal(content, s.conf.PSKey)

	w.WriteHeader(http.StatusOK)
	w.Write(respData)
}

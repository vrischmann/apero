package main

import (
	"net/http"

	"github.com/vrischmann/hutil/v2"
)

type uiHandler struct {
	conf serverConfig
}

func newUIHandler(conf serverConfig) *uiHandler {
	return &uiHandler{
		conf: conf,
	}
}

func (s *uiHandler) handle(w http.ResponseWriter, req *http.Request, tail string) {
	head, tail := hutil.ShiftPath(tail)
	switch head {
	case "":
		s.handleIndex(w, req)
	}
}

func (s *uiHandler) handleIndex(w http.ResponseWriter, req *http.Request) {

}

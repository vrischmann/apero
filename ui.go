package main

import (
	"net/http"

	"github.com/vrischmann/hutil/v2"
)

type uiServer struct {
}

func newUIServer() *uiServer {
	return &uiServer{}
}

func (s *uiServer) handle(w http.ResponseWriter, req *http.Request, tail string) {
	head, tail := hutil.ShiftPath(tail)
	switch head {
	case "":
		s.handleIndex(w, req)
	}
}

func (s *uiServer) handleIndex(w http.ResponseWriter, req *http.Request) {

}

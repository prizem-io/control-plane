// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type (
	Server struct {
		router *mux.Router
	}
	Variables func(r *http.Request) map[string]string

	Error struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}
)

func NewServer() *Server {
	return &Server{
		router: mux.NewRouter(),
	}
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (e *Error) Error() string {
	return e.Message
}

func sendError(w http.ResponseWriter, status int, message string) {
	err := Error{
		Status:  status,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&err)
}

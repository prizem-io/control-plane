// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prizem-io/api/v1"
)

func (s *Server) RegisterRoutes(service api.Routes) {
	routes := Routes{
		vars:    mux.Vars,
		service: service,
	}
	s.router.HandleFunc("/v1/routes", routes.GetServices).Methods("GET")
	s.router.HandleFunc("/v1/routes", routes.SetService).Methods("POST", "PUT")
	s.router.HandleFunc("/v1/routes/{serviceName}", routes.RemoveService).Methods("DELETE")
}

type Routes struct {
	vars    Variables
	service api.Routes
}

func NewRoutes(vars Variables, service api.Routes) *Routes {
	return &Routes{
		vars:    vars,
		service: service,
	}
}

func (s *Routes) GetServices(w http.ResponseWriter, r *http.Request) {
	currentIndex, _ := strconv.ParseInt(r.Header.Get("If-None-Match"), 10, 63)

	services, index, useCache, err := s.service.GetServices(currentIndex)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if useCache {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("ETag", fmt.Sprintf("%d", index))
	json.NewEncoder(w).Encode(&api.RouteInfo{
		Services: services,
	})
}

func (s *Routes) SetService(w http.ResponseWriter, r *http.Request) {
	var service api.Service
	err := json.NewDecoder(r.Body).Decode(&service)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	index, err := s.service.SetService(service)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("ETag", fmt.Sprintf("%d", index))
	w.WriteHeader(http.StatusCreated)
}

func (s *Routes) RemoveService(w http.ResponseWriter, r *http.Request) {
	vars := s.vars(r)
	serviceName := vars["serviceName"]

	index, err := s.service.RemoveService(serviceName)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("ETag", fmt.Sprintf("%d", index))
}

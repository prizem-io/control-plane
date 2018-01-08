// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package rest

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/prizem-io/api/v1"
	"github.com/satori/go.uuid"
)

func (s *Server) RegisterEndpoints(service api.Endpoints) {
	endpoints := Endpoints{
		vars:    mux.Vars,
		service: service,
	}
	s.router.HandleFunc("/v1/endpoints", endpoints.GetEndpoints).Methods("GET")
	s.router.HandleFunc("/v1/endpoints", endpoints.AddEndpoints).Methods("POST")
	s.router.HandleFunc("/v1/endpoints/{nodeID}", endpoints.RemoveNodeEndpoints).Methods("DELETE")
	s.router.HandleFunc("/v1/endpoints/{nodeID}/{serviceNames}", endpoints.RemoveServiceEndpoints).Methods("DELETE")
}

type Endpoints struct {
	vars    Variables
	service api.Endpoints
}

func NewEndpoints(vars Variables, service api.Endpoints) *Endpoints {
	return &Endpoints{
		vars:    vars,
		service: service,
	}
}

func (s *Endpoints) GetEndpoints(w http.ResponseWriter, r *http.Request) {
	currentIndex, _ := strconv.ParseInt(r.Header.Get("If-None-Match"), 10, 63)

	nodes, index, useCache, err := s.service.GetEndpoints(currentIndex)
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
	json.NewEncoder(w).Encode(&api.EndpointInfo{
		Nodes: nodes,
	})
}

func (s *Endpoints) AddEndpoints(w http.ResponseWriter, r *http.Request) {
	var node api.Node
	err := json.NewDecoder(r.Body).Decode(&node)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If not provided, extract the address/IP from the remote address
	if node.Address == nil {
		addr := r.RemoteAddr
		i := strings.LastIndex(addr, ":")
		if i != -1 {
			addr = addr[:i]
		}
		// Remove IPv6 surrounding brackets
		addr = strings.TrimPrefix(addr, "[")
		addr = strings.TrimSuffix(addr, "]")
		if addr == "::1" {
			addr = "127.0.0.1"
		}
		node.Address = net.ParseIP(addr)
	}

	if node.Address == nil {
		sendError(w, http.StatusBadRequest, "address is required")
		return
	}

	for i := range node.Services {
		if uuid.Equal(node.Services[i].ID, uuid.Nil) {
			node.Services[i].ID = uuid.NewV4()
		}
	}

	modification, err := s.service.AddEndpoints(node)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("ETag", fmt.Sprintf("%d", modification.Index))
	if len(modification.Services) > 0 {
		w.WriteHeader(http.StatusCreated)
	}
	json.NewEncoder(w).Encode(modification)
}

func (s *Endpoints) RemoveNodeEndpoints(w http.ResponseWriter, r *http.Request) {
	vars := s.vars(r)
	nodeID := vars["nodeID"]

	modification, err := s.service.RemoveEndpoints(nodeID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("ETag", fmt.Sprintf("%d", modification.Index))
	json.NewEncoder(w).Encode(modification)
}

func (s *Endpoints) RemoveServiceEndpoints(w http.ResponseWriter, r *http.Request) {
	vars := s.vars(r)
	nodeID := vars["nodeID"]
	serviceNamesStr := vars["serviceNames"]
	serviceNames := strings.Split(serviceNamesStr, ",")

	modification, err := s.service.RemoveEndpoints(nodeID, serviceNames...)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("ETag", fmt.Sprintf("%d", modification.Index))
	json.NewEncoder(w).Encode(modification)
}

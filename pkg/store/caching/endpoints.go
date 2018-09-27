// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package caching

import (
	"sync"

	"github.com/prizem-io/api/v1"
)

type (
	EndpointsCallback func(version int64, nodes []api.Node) error

	Endpoints struct {
		service api.Endpoints

		mu      sync.RWMutex
		version int64
		nodes   []api.Node

		callbacks []EndpointsCallback
	}
)

func NewEndpoints(service api.Endpoints) *Endpoints {
	return &Endpoints{
		service: service,
	}
}

func (e *Endpoints) AddCallback(callback EndpointsCallback) {
	e.callbacks = append(e.callbacks, callback)
}

func (e *Endpoints) SetEndpoints(version int64, nodes []api.Node) {
	e.mu.Lock()
	if version > e.version {
		e.nodes = nodes
		e.version = version
	}
	e.mu.Unlock()
}

func (e *Endpoints) GetEndpoints(currentVersion int64) (nodes []api.Node, version int64, useCache bool, err error) {
	e.mu.RLock()
	nodes = e.nodes
	version = e.version
	e.mu.RUnlock()

	if nodes != nil {
		if currentVersion != 0 && currentVersion == version {
			useCache = true
			return
		}
	}

	nodes, version, useCache, err = e.service.GetEndpoints(0) // Pass 0 to force load
	if currentVersion != 0 && currentVersion == version {
		useCache = true
	}

	e.SetEndpoints(version, nodes)

	return
}

func (e *Endpoints) AddEndpoints(node api.Node) (modification *api.Modification, err error) {
	modification, err = e.service.AddEndpoints(node)
	if err != nil {
		return nil, err
	}
	if modification.Node || len(modification.Services) > 0 {
		e.publishNewEndpoints()
	}
	return
}

func (e *Endpoints) RemoveEndpoints(nodeID string, serviceNames ...string) (modification *api.Modification, err error) {
	modification, err = e.service.RemoveEndpoints(nodeID, serviceNames...)
	if err != nil {
		return nil, err
	}
	if modification.Node || len(modification.Services) > 0 {
		e.publishNewEndpoints()
	}
	return
}

func (e *Endpoints) publishNewEndpoints() {
	nodes, version, _, err := e.GetEndpoints(0)
	if err != nil {
		return
	}
	for _, cb := range e.callbacks {
		cb(version, nodes)
	}
}

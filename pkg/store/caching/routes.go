// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package caching

import (
	"sync"

	"github.com/prizem-io/api/v1"
)

type (
	RoutesCallback func(version int64, nodes []api.Service) error

	Routes struct {
		service api.Routes

		mu       sync.RWMutex
		version  int64
		services []api.Service

		callbacks []RoutesCallback
	}
)

func NewRoutes(service api.Routes) *Routes {
	return &Routes{
		service: service,
	}
}

func (r *Routes) AddCallback(callback RoutesCallback) {
	r.callbacks = append(r.callbacks, callback)
}

func (r *Routes) SetServices(version int64, services []api.Service) {
	r.mu.Lock()
	if version > r.version {
		r.services = services
		r.version = version
	}
	r.mu.Unlock()
}

func (r *Routes) GetServices(currentVersion int64) (services []api.Service, version int64, useCache bool, err error) {
	r.mu.RLock()
	services = r.services
	version = r.version
	r.mu.RUnlock()

	if services != nil {
		if currentVersion != 0 && currentVersion == version {
			useCache = true
			return
		}
	}

	services, version, useCache, err = r.service.GetServices(0) // Pass 0 to force load
	if currentVersion == version {
		useCache = true
	}

	r.SetServices(version, services)

	return
}

func (r *Routes) SetService(service api.Service) (index int64, err error) {
	index, err = r.service.SetService(service)
	r.publishNewRoutes()
	return
}

func (r *Routes) RemoveService(serviceName string) (index int64, err error) {
	index, err = r.service.RemoveService(serviceName)
	r.publishNewRoutes()
	return
}

func (r *Routes) publishNewRoutes() {
	services, version, _, err := r.GetServices(0)
	if err != nil {
		return
	}
	for _, cb := range r.callbacks {
		cb(version, services)
	}
}

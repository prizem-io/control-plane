// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package logging

import (
	"github.com/prizem-io/api/v1"
)

type Routing struct {
	service api.Routes
	logger  logger
}

func NewRouting(service api.Routes, logger logger) *Routing {
	return &Routing{
		service: service,
		logger:  logger,
	}
}

func (r *Routing) GetServices(currentIndex int64) (services []api.Service, index int64, useCache bool, err error) {
	defer func() {
		if err != nil {
			r.logger.Error(err)
		}
	}()
	return r.service.GetServices(currentIndex)
}

func (r *Routing) SetService(service api.Service) (index int64, err error) {
	defer func() {
		if err != nil {
			r.logger.Error(err)
		}
	}()
	return r.service.SetService(service)
}

func (r *Routing) RemoveService(serviceName string) (index int64, err error) {
	defer func() {
		if err != nil {
			r.logger.Error(err)
		}
	}()
	return r.service.RemoveService(serviceName)
}

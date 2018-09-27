// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package logging

import (
	"github.com/prizem-io/api/v1"
	"github.com/prizem-io/control-plane/pkg/log"
)

type Endpoints struct {
	service api.Endpoints
	logger  log.Logger
}

func NewEndpoints(service api.Endpoints, logger log.Logger) *Endpoints {
	return &Endpoints{
		service: service,
		logger:  logger,
	}
}

func (e *Endpoints) GetEndpoints(currentIndex int64) (nodes []api.Node, index int64, useCache bool, err error) {
	defer func() {
		if err != nil {
			e.logger.Errorf("error retrieving endpoints: %v", err)
		}
	}()
	return e.service.GetEndpoints(currentIndex)
}

func (e *Endpoints) AddEndpoints(node api.Node) (modification *api.Modification, err error) {
	defer func() {
		if err != nil {
			e.logger.Errorf("error adding endpoints: %v", err)
		}
	}()
	return e.service.AddEndpoints(node)
}

func (e *Endpoints) RemoveEndpoints(nodeID string, serviceNames ...string) (modification *api.Modification, err error) {
	defer func() {
		if err != nil {
			e.logger.Errorf("error removing endpoints: %v", err)
		}
	}()
	return e.service.RemoveEndpoints(nodeID, serviceNames...)
}

// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grpc

import (
	"context"
	"io"
	"sync"

	"github.com/prizem-io/api/v1"
	"github.com/prizem-io/api/v1/convert"
	"github.com/prizem-io/api/v1/proto"

	"github.com/prizem-io/control-plane/pkg/log"
)

type Routes struct {
	logger  log.Logger
	service api.Routes

	subscriptionsMu sync.RWMutex
	subscriptions   map[string]proto.RouteDiscovery_StreamRoutesServer
}

func NewRoutes(logger log.Logger, service api.Routes) *Routes {
	return &Routes{
		logger:        logger,
		service:       service,
		subscriptions: make(map[string]proto.RouteDiscovery_StreamRoutesServer, 100),
	}
}

func (s *Routes) GetRoutes(_ context.Context, request *proto.RoutesRequest) (*proto.RoutesCatalog, error) {
	services, index, useCache, err := s.service.GetServices(request.Version)
	if err != nil {
		return nil, err
	}

	return &proto.RoutesCatalog{
		UseCache: useCache,
		Version:  index,
		Services: convert.EncodeServices(services),
	}, nil
}

func (s *Routes) StreamRoutes(stream proto.RouteDiscovery_StreamRoutesServer) error {
	var nodeID string

	for {
		request, err := stream.Recv()
		if err != nil {
			if nodeID != "" {
				s.logger.Debugf("Removing subscription for %s", nodeID)
				s.subscriptionsMu.Lock()
				delete(s.subscriptions, nodeID)
				s.subscriptionsMu.Unlock()
			}
			if err == io.EOF {
				return nil
			}
			return err
		}

		if nodeID == "" && request.NodeID != "" {
			s.logger.Debugf("Adding routes subscription for %s", request.NodeID)
			s.subscriptionsMu.Lock()
			s.subscriptions[request.NodeID] = stream
			s.subscriptionsMu.Unlock()
			nodeID = request.NodeID
		}

		services, index, useCache, err := s.service.GetServices(request.Version)
		if err != nil {
			return err
		}

		stream.Send(&proto.RoutesCatalog{
			UseCache: useCache,
			Version:  index,
			Services: convert.EncodeServices(services),
		})
	}
}

func (s *Routes) PublishRoutes(version int64, services []api.Service) error {
	s.logger.Debug("Publishing routes to gRPC clients")
	catalog := proto.RoutesCatalog{
		UseCache: false,
		Version:  version,
		Services: convert.EncodeServices(services),
	}

	type nodeIDStreamPair struct {
		nodeID string
		stream proto.RouteDiscovery_StreamRoutesServer
	}

	pairs := make([]nodeIDStreamPair, len(s.subscriptions))
	var i int
	s.subscriptionsMu.RLock()
	for nodeID, stream := range s.subscriptions {
		pairs[i] = nodeIDStreamPair{nodeID, stream}
		i++
	}
	s.subscriptionsMu.RUnlock()

	for _, pair := range pairs {
		s.logger.Debugf("Sending routes catalog update to %s", pair.nodeID)
		err := pair.stream.Send(&catalog)
		if err != nil {
			s.logger.Errorf("Error sending routes catalog to client: %v", err)
		}
	}

	return nil
}

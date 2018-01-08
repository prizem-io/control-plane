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
	log "github.com/sirupsen/logrus"
)

type Routes struct {
	service api.Routes

	subscriptionsMu sync.RWMutex
	subscriptions   map[string]proto.RouteDiscovery_StreamRoutesServer
}

func NewRoutes(service api.Routes) *Routes {
	return &Routes{
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
				log.Debugf("Removing subscription for %s", nodeID)
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
			log.Debugf("Adding routes subscription for %s", request.NodeID)
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
	log.Debug("Publishing routes to gRPC clients")
	catalog := proto.RoutesCatalog{
		UseCache: false,
		Version:  version,
		Services: convert.EncodeServices(services),
	}

	s.subscriptionsMu.RLock()
	for nodeID, stream := range s.subscriptions {
		log.Debugf("Sending routes catalog update to %s", nodeID)
		err := stream.Send(&catalog) // Is this thread-safe?
		if err != nil {
			// TODO
		}
	}
	s.subscriptionsMu.RUnlock()
	return nil
}

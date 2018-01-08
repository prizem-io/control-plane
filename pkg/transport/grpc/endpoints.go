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

type Endpoints struct {
	service api.Endpoints

	subscriptionsMu sync.RWMutex
	subscriptions   map[string]proto.EndpointDiscovery_StreamEndpointsServer
}

func NewEndpoints(service api.Endpoints) *Endpoints {
	return &Endpoints{
		service:       service,
		subscriptions: make(map[string]proto.EndpointDiscovery_StreamEndpointsServer, 100),
	}
}

func (s *Endpoints) GetEndpoints(_ context.Context, request *proto.EndpointsRequest) (*proto.EndpointsCatalog, error) {
	nodes, index, useCache, err := s.service.GetEndpoints(request.Version)
	if err != nil {
		return nil, err
	}

	return &proto.EndpointsCatalog{
		UseCache: useCache,
		Version:  index,
		Nodes:    convert.EncodeNodes(nodes),
	}, nil
}

func (s *Endpoints) StreamEndpoints(stream proto.EndpointDiscovery_StreamEndpointsServer) error {
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
			log.Debugf("Adding endpoints subscription for %s", request.NodeID)
			s.subscriptionsMu.Lock()
			s.subscriptions[request.NodeID] = stream
			s.subscriptionsMu.Unlock()
			nodeID = request.NodeID
		}

		nodes, index, useCache, err := s.service.GetEndpoints(request.Version)
		if err != nil {
			return err
		}

		stream.Send(&proto.EndpointsCatalog{
			UseCache: useCache,
			Version:  index,
			Nodes:    convert.EncodeNodes(nodes),
		})
	}
}

func (s *Endpoints) PublishEndpoints(version int64, nodes []api.Node) error {
	log.Debug("Publishing endpoints to gRPC clients")
	catalog := proto.EndpointsCatalog{
		UseCache: false,
		Version:  version,
		Nodes:    convert.EncodeNodes(nodes),
	}

	s.subscriptionsMu.RLock()
	for nodeID, stream := range s.subscriptions {
		log.Debugf("Sending endpoints catalog update to %s", nodeID)
		stream.Send(&catalog) // Is this thread-safe?
	}
	s.subscriptionsMu.RUnlock()
	return nil
}

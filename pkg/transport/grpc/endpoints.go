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

type Endpoints struct {
	logger  log.Logger
	service api.Endpoints

	subscriptionsMu sync.RWMutex
	subscriptions   map[string]proto.EndpointDiscovery_StreamEndpointsServer
}

func NewEndpoints(logger log.Logger, service api.Endpoints) *Endpoints {
	return &Endpoints{
		logger:        logger,
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
			s.logger.Debugf("Adding endpoints subscription for %s", request.NodeID)
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
	s.logger.Debug("Publishing endpoints to gRPC clients")
	catalog := proto.EndpointsCatalog{
		UseCache: false,
		Version:  version,
		Nodes:    convert.EncodeNodes(nodes),
	}

	type nodeIDStreamPair struct {
		nodeID string
		stream proto.EndpointDiscovery_StreamEndpointsServer
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
		s.logger.Debugf("Sending endpoints catalog update to %s", pair.nodeID)
		err := pair.stream.Send(&catalog)
		if err != nil {
			s.logger.Errorf("Error sending endpoints catalog to client: %v", err)
		}
	}

	return nil
}

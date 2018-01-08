// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package healthcheck

import (
	"encoding/json"
	"io"
	"sync"
	"sync/atomic"
	"unsafe"

	nats "github.com/nats-io/go-nats"
	"github.com/prizem-io/api/v1"
	"github.com/prizem-io/api/v1/proto"
	uuid "github.com/satori/go.uuid"
)

type Service struct {
	endpoints api.Endpoints
	conn      *nats.Conn
	sub       *nats.Subscription

	nodes unsafe.Pointer

	mu      sync.Mutex
	streams map[string]proto.HealthCheckReporter_StreamHealthCheckServer
}

type message struct {
	ServerID string          `json:"serverId"`
	Type     string          `json:"type"`
	Data     json.RawMessage `json:"data"`
}

func New(endpoints api.Endpoints) *Service {
	return &Service{
		endpoints: endpoints,
	}
}

func (s *Service) Subscribe() error {
	var err error
	s.sub, err = s.conn.Subscribe("routes", func(m *nats.Msg) {
		var msg message
		err := json.Unmarshal(m.Data, &msg)
		if err != nil {
			return
		}

		if msg.Type == "invalidate" {
			nodes, _, _, err := s.endpoints.GetEndpoints(0)
			if err != nil {
				return
			}

			atomic.StorePointer(&s.nodes, unsafe.Pointer(&nodes))
		}
	})

	return err
}

func (s *Service) StreamHealthCheck(stream proto.HealthCheckReporter_StreamHealthCheckServer) error {
	var nodeID string
	defer func() {
		if nodeID != "" {
			s.mu.Lock()
			delete(s.streams, nodeID)
			s.mu.Unlock()
		}
	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch v := in.HealthRequest.(type) {
		case *proto.HeathRequest_Initialization:
			if nodeID != "" {
				continue
			}

			nodeID = v.Initialization.NodeId
			s.mu.Lock()
			s.streams[nodeID] = stream
			s.mu.Unlock()

			s.sendAssignments(nodeID, stream)
		case *proto.HeathRequest_Status:

		}
	}
}

func (s *Service) sendAssignments(nodeID string, stream proto.HealthCheckReporter_StreamHealthCheckServer) error {
	ptr := atomic.LoadPointer(&s.nodes)
	nodes := *(*[]api.Node)(ptr)

	nodeUUID, err := uuid.FromString(nodeID)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		if n.ID == nodeUUID {
			assignments := make([]proto.HealthAssignment, len(n.Services))
			for i, s := range n.Services {
				assignments[i] = proto.HealthAssignment{
					NodeID:  nodeID,
					Service: s.Name,
				}
			}
			err := stream.Send(&proto.HealthAssignments{
				Assignments: assignments,
			})
			return err
		}
	}

	return nil
}

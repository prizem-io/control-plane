// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package nats

import (
	"github.com/gogo/protobuf/proto"
	nats "github.com/nats-io/go-nats"
	"github.com/pkg/errors"
	"github.com/prizem-io/api/v1"
	"github.com/prizem-io/api/v1/convert"
	pb "github.com/prizem-io/api/v1/proto"

	"github.com/prizem-io/control-plane/pkg/log"
)

const endpointsSubject = "endpoints"

type (
	EndpointsCallback func(version int64, nodes []api.Node)

	Endpoints struct {
		logger   log.Logger
		serverID string

		conn *nats.Conn
		sub  *nats.Subscription

		callbacks []EndpointsCallback
	}
)

func NewEndpoints(logger log.Logger, serverID string, conn *nats.Conn) *Endpoints {
	return &Endpoints{
		logger:   logger,
		serverID: serverID,
		conn:     conn,
	}
}

func (e *Endpoints) PublishEndpoints(version int64, nodes []api.Node) error {
	e.logger.Debugf("Replicating endpoints version %d", version)

	data, err := proto.Marshal(&pb.EndpointsCatalog{
		Version: version,
		Nodes:   convert.EncodeNodes(nodes),
	})
	if err != nil {
		e.logger.Errorf("Error marshalling protobuf: %v", err)
	}
	msg := pb.Message{
		Type:     "replicate",
		ServerID: e.serverID,
		Data:     data,
	}
	data, err = proto.Marshal(&msg)
	if err != nil {
		return errors.Wrap(err, "Could not marshal routing message")
	}
	err = e.conn.Publish(endpointsSubject, data)
	if err != nil {
		return errors.Wrap(err, "Could not publish routing message")
	}
	return nil
}

func (e *Endpoints) AddCallback(callback EndpointsCallback) {
	e.callbacks = append(e.callbacks, callback)
}

func (e *Endpoints) Subscribe() error {
	if e.sub == nil {
		var err error
		e.sub, err = e.conn.Subscribe(endpointsSubject, func(m *nats.Msg) {
			var msg pb.Message
			err := proto.Unmarshal(m.Data, &msg)
			if err != nil {
				e.logger.Errorf("could not unmarshal control plane message: %v", err)
				return
			}

			if msg.ServerID == e.serverID {
				return
			}

			if msg.Type == "replicate" {
				var catalog pb.EndpointsCatalog
				var version int64
				var nodes []api.Node
				hasData := len(msg.Data) > 0

				if hasData {
					err := proto.Unmarshal(msg.Data, &catalog)
					if err != nil {
						e.logger.Errorf("could not unmarshal replicate endpoints message: %v", err)
					} else {
						version = catalog.Version
						nodes = convert.DecodeNodes(catalog.Nodes)

						for _, cb := range e.callbacks {
							cb(version, nodes)
						}
					}
				} else {
					for _, cb := range e.callbacks {
						cb(version, nil)
					}
				}
			}
		})
		if err != nil {
			return errors.Wrap(err, "Could not subscribe to routing subject")
		}
	}

	return nil
}

func (e *Endpoints) Unsubscribe() error {
	if e.sub != nil {
		return e.sub.Unsubscribe()
	}

	return nil
}

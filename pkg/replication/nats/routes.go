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

const routesSubject = "routes"

type (
	RoutesCallback func(version int64, services []api.Service)

	Routes struct {
		logger   log.Logger
		serverID string

		conn *nats.Conn
		sub  *nats.Subscription

		callbacks []RoutesCallback
	}
)

func NewRoutes(logger log.Logger, serverID string, conn *nats.Conn) *Routes {
	return &Routes{
		logger:   logger,
		serverID: serverID,
		conn:     conn,
	}
}

func (r *Routes) PublishRoutes(version int64, services []api.Service) error {
	r.logger.Debugf("Replicating routes version %d", version)

	data, err := proto.Marshal(&pb.RoutesCatalog{
		Version:  version,
		Services: convert.EncodeServices(services),
	})
	if err != nil {
		r.logger.Errorf("Error marshalling protobuf: %v", err)
	}
	msg := pb.Message{
		Type:     "replicate",
		ServerID: r.serverID,
		Data:     data,
	}
	data, err = proto.Marshal(&msg)
	if err != nil {
		return errors.Wrap(err, "Could not marshal routing message")
	}
	err = r.conn.Publish(routesSubject, data)
	if err != nil {
		return errors.Wrap(err, "Could not publish routing message")
	}
	return nil
}

func (r *Routes) AddCallback(callback RoutesCallback) {
	r.callbacks = append(r.callbacks, callback)
}

func (r *Routes) Subscribe() error {
	if r.sub == nil {
		var err error
		r.sub, err = r.conn.Subscribe(routesSubject, func(m *nats.Msg) {
			var msg pb.Message
			err := proto.Unmarshal(m.Data, &msg)
			if err != nil {
				r.logger.Errorf("could not unmarshal control plane message: %v", err)
				return
			}

			if msg.ServerID == r.serverID {
				return
			}

			if msg.Type == "replicate" {
				var catalog pb.RoutesCatalog
				var version int64
				var services []api.Service
				hasData := len(msg.Data) > 0

				if hasData {
					err := proto.Unmarshal(msg.Data, &catalog)
					if err != nil {
						r.logger.Errorf("could not unmarshal replicate routes message: %v", err)
					} else {
						version = catalog.Version
						services = convert.DecodeServices(catalog.Services)

						for _, cb := range r.callbacks {
							cb(version, services)
						}
					}
				} else {
					for _, cb := range r.callbacks {
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

func (r *Routes) Unsubscribe() error {
	if r.sub != nil {
		return r.sub.Unsubscribe()
	}

	return nil
}

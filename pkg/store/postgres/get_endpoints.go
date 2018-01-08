// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"database/sql"
	"encoding/json"
	"net"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/prizem-io/api/v1"
	uuid "github.com/satori/go.uuid"
)

type dbNode struct {
	NodeID     uuid.UUID       `db:"node_id"`
	Geography  string          `db:"geography"`
	Datacenter string          `db:"datacenter"`
	Address    string          `db:"address"`
	Metadata   json.RawMessage `db:"metadata"`
}

type dbEndpoint struct {
	EndpointID  uuid.UUID       `db:"endpoint_id"`
	ServiceName string          `db:"service_name"`
	Config      json.RawMessage `db:"config"`
}

func (s *Store) GetEndpoints(currentIndex int64) (nodes []api.Node, index int64, useCache bool, err error) {
	var tx *sqlx.Tx
	tx, err = s.db.Beginx()
	if err != nil {
		err = errors.Wrap(err, "Could not begin transaction")
		return
	}
	defer s.handleDeferTx(tx, err)

	index = 1
	err = tx.Get(&index, "SELECT index FROM source WHERE name = 'endpoints'")
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			return
		}
	}

	if currentIndex == index {
		useCache = true
		return
	}

	var dbNodes []dbNode
	err = tx.Select(&dbNodes, "SELECT node_id, geography, datacenter, address, metadata FROM node")
	if err != nil {
		return
	}

	nodes = make([]api.Node, len(dbNodes))
	for i, n := range dbNodes {
		var dbServices []dbEndpoint
		err = tx.Select(&dbServices, "SELECT endpoint_id, service_name, config FROM endpoint WHERE node_id = $1", n.NodeID)
		if err != nil {
			return
		}

		services := make([]api.ServiceInstance, len(dbServices))
		for i, s := range dbServices {
			var si api.ServiceInstance
			err = json.Unmarshal(s.Config, &si)
			if err != nil {
				return
			}

			si.ID = s.EndpointID
			si.Service = s.ServiceName
			services[i] = si
		}

		var metadata map[string]interface{}
		err = json.Unmarshal(n.Metadata, &metadata)
		if err != nil {
			return
		}

		nodes[i] = api.Node{
			ID:         n.NodeID,
			Geography:  n.Geography,
			Datacenter: n.Datacenter,
			Address:    net.ParseIP(n.Address),
			Metadata:   metadata,
			Services:   services,
		}
	}

	return
}

// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"database/sql"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/prizem-io/api/v1"
)

func (s *Store) AddEndpoints(node api.Node) (modification *api.Modification, err error) {
	var tx *sqlx.Tx
	tx, err = s.db.Beginx()
	if err != nil {
		err = errors.Wrap(err, "could not begin transaction")
		return
	}
	defer s.handleDeferTx(tx, err)

	var currentIndex int64
	err = tx.Get(&currentIndex, "SELECT index FROM source WHERE name = 'endpoints' FOR UPDATE")
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "could not select current endpoints index")
			return
		}
	}

	var nodeExists bool
	err = tx.Get(&nodeExists, "SELECT EXISTS (SELECT 1 FROM node WHERE node_id = $1 LIMIT 1)", node.ID)
	if err != nil {
		err = errors.Wrap(err, "could not determine if node exists")
		return
	}

	var data []byte
	data, err = json.Marshal(node.Metadata)
	if err != nil {
		err = errors.Wrap(err, "could not marshal node metadata")
		return
	}

	if !nodeExists {
		_, err = tx.Exec("INSERT INTO node (node_id, geography, datacenter, address, metadata) VALUES ($1, $2, $3, $4, $5)",
			node.ID, node.Geography, node.Datacenter, node.Address.String(), json.RawMessage(data))
		if err != nil {
			err = errors.Wrapf(err, "could not insert node %s", node.ID)
			return
		}
	} else {
		_, err = tx.Exec("UPDATE node SET geography = $1, datacenter = $2, address = $3, metadata = $4 WHERE node_id = $5",
			node.Geography, node.Datacenter, node.Address.String(), json.RawMessage(data), node.ID)
		if err != nil {
			err = errors.Wrapf(err, "could not update node %s", node.ID)
			return
		}
	}

	services := make([]string, 0, len(node.Services))

	for _, service := range node.Services {
		var data []byte
		data, err = json.Marshal(service)
		if err != nil {
			return
		}

		var endpointExists bool
		err = tx.Get(&endpointExists, "SELECT EXISTS (SELECT 1 FROM endpoint WHERE endpoint_id = $1)", service.ID)
		if err != nil {
			err = errors.Wrap(err, "could not determine if service exists")
			return
		}

		if !endpointExists {
			_, err = tx.Exec("INSERT INTO endpoint (endpoint_id, node_id, service_name, config) VALUES ($1, $2, $3, $4)",
				service.ID, node.ID, service.Service, json.RawMessage(data))
			if err != nil {
				err = errors.Wrapf(err, "could not insert endpoint for service %s", service.Name)
				return
			}

			services = append(services, service.Name)
		} else {
			_, err = tx.Exec("UPDATE endpoint SET node_id = $1, service_name = $2, config = $3 WHERE endpoint_id = $4",
				node.ID, service.Service, json.RawMessage(data), service.ID)
			if err != nil {
				err = errors.Wrapf(err, "could not update endpoint for service %s", service.Name)
				return
			}

			services = append(services, service.Name)
		}
	}

	newIndex := currentIndex

	if !nodeExists || len(services) > 0 {
		newIndex++
		if newIndex == 0 {
			newIndex++
		}
		err = updateSource(tx, newIndex, "endpoints")
		if err != nil {
			return
		}
	}

	modification = &api.Modification{
		Index:    newIndex,
		Node:     !nodeExists,
		Services: services,
	}

	return
}

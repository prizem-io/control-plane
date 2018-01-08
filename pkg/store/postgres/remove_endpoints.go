// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/prizem-io/api/v1"
)

func (s *Store) RemoveEndpoints(nodeID string, serviceNames ...string) (modification *api.Modification, err error) {
	var tx *sqlx.Tx
	tx, err = s.db.Beginx()
	if err != nil {
		err = errors.Wrap(err, "Could not begin transaction")
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

	var services []string
	var nodeDeleted bool

	if len(serviceNames) > 0 {
		services = make([]string, 0, len(serviceNames))

		for _, serviceName := range serviceNames {
			var result sql.Result
			result, err = tx.Exec("DELETE FROM endpoint WHERE node_id = $1 AND service_name = $2", nodeID, serviceName)
			if err != nil {
				err = errors.Wrapf(err, "Could delete service %s from node %s", serviceName, nodeID)
				return
			}

			var affected int64
			affected, err = result.RowsAffected()
			if err != nil {
				err = errors.Wrapf(err, "Could not read result of clean up for node %s", nodeID)
				return
			}

			if affected > 0 {
				services = append(services, serviceName)
			}

			var nodeExists bool
			err = tx.Get(&nodeExists, "SELECT EXISTS (SELECT 1 FROM endpoint WHERE node_id = $1 LIMIT 1)", nodeID)
			if err != nil {
				err = errors.Wrapf(err, "Could not determine if endpoints for node %s remain", nodeID)
				return
			}

			if !nodeExists {
				result, err = tx.Exec("DELETE FROM node WHERE node_id = $1", nodeID)
				if err != nil {
					err = errors.Wrapf(err, "Could not clean up node %s", nodeID)
					return
				}
				affected, err = result.RowsAffected()
				if err != nil {
					err = errors.Wrapf(err, "Could not read result of clean up for node %s", nodeID)
					return
				}
				nodeDeleted = affected > 0
			}
		}
	} else {
		err = tx.Select(&services, "SELECT service_name FROM endpoint WHERE node_id = $1", nodeID)
		if err != nil {
			err = errors.Wrapf(err, "Could not select existing endpoints for node %s", nodeID)
			return
		}

		if len(services) > 0 {
			_, err = tx.Exec("DELETE FROM endpoint WHERE node_id = $1", nodeID)
			if err != nil {
				err = errors.Wrapf(err, "Could not clean up endpoints for node %s", nodeID)
				return
			}
		}

		var result sql.Result
		var affected int64

		result, err = tx.Exec("DELETE FROM node WHERE node_id = $1", nodeID)
		if err != nil {
			err = errors.Wrapf(err, "Could not clean up node %s", nodeID)
			return
		}

		affected, err = result.RowsAffected()
		if err != nil {
			err = errors.Wrapf(err, "Could not read result of clean up for node %s", nodeID)
			return
		}
		nodeDeleted = affected > 0
	}

	newIndex := currentIndex

	if nodeDeleted || len(services) > 0 {
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
		Node:     nodeDeleted,
		Services: services,
	}

	return
}

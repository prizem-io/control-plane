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

func (s *Store) SetService(service api.Service) (index int64, err error) {
	var tx *sqlx.Tx
	tx, err = s.db.Beginx()
	if err != nil {
		err = errors.Wrap(err, "Could not begin transaction")
		return
	}
	defer s.handleDeferTx(tx, err)

	index = 1
	err = tx.Get(&index, "SELECT index FROM source WHERE name = 'routing' FOR UPDATE")
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "could not select current routing index")
			return
		}
	}

	var data []byte
	data, err = json.Marshal(&service)
	if err != nil {
		err = errors.Wrapf(err, "Could not marshal service %s", service.Name)
		return
	}

	var result sql.Result
	result, err = tx.Exec("UPDATE service SET config = $1 WHERE service_name = $2",
		json.RawMessage(data), service.Name)
	if err != nil {
		err = errors.Wrapf(err, "Could not update service %s", service.Name)
		return
	}

	var affected int64
	affected, err = result.RowsAffected()
	if err != nil {
		err = errors.Wrapf(err, "Could not determine if service %s was updated", service.Name)
		return
	}

	if affected == 0 {
		_, err = tx.Exec("INSERT INTO service (service_name, config) VALUES ($1, $2)",
			service.Name, json.RawMessage(data))
		if err != nil {
			err = errors.Wrapf(err, "Could not update insert service %s", service.Name)
			return
		}
	}

	index++
	if index == 0 {
		index++
	}
	err = updateSource(tx, index, "routing")
	if err != nil {
		return
	}

	return
}

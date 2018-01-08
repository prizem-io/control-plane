// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func (s *Store) RemoveService(serviceName string) (index int64, err error) {
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

	var result sql.Result
	result, err = tx.Exec("DELETE FROM service WHERE service_name = $1", serviceName)
	if err != nil {
		err = errors.Wrapf(err, "Could not delete service %s", serviceName)
		return
	}

	var affected int64
	affected, err = result.RowsAffected()
	if err != nil {
		err = errors.Wrapf(err, "Could not determine if service %s was removed", serviceName)
		return
	}

	if affected > 0 {
		index++
		if index == 0 {
			index++
		}
		err = updateSource(tx, index, "routing")
		if err != nil {
			return
		}
	}

	return
}

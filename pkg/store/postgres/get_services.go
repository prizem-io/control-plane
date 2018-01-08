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

type dbService struct {
	ServiceName string          `db:"service_name"`
	Config      json.RawMessage `db:"config"`
}

func (s *Store) GetServices(currentIndex int64) (services []api.Service, index int64, useCache bool, err error) {
	var tx *sqlx.Tx
	tx, err = s.db.Beginx()
	if err != nil {
		err = errors.Wrap(err, "Could not begin transaction")
		return
	}
	defer s.handleDeferTx(tx, err)

	index = 1
	err = tx.Get(&index, "SELECT index FROM source WHERE name = 'routing'")
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

	var dbServices []dbService
	err = tx.Select(&dbServices, "SELECT config FROM service ORDER BY service_name")
	if err != nil {
		return
	}

	services = make([]api.Service, len(dbServices))
	for i, s := range dbServices {
		err = json.Unmarshal(s.Config, &services[i])
		if err != nil {
			return
		}
	}

	return
}

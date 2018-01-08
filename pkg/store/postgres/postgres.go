// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Store struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) handleDeferTx(tx *sqlx.Tx, err error) {
	if err != nil {
		_ = tx.Rollback()
	} else {
		_ = tx.Commit()
	}
}

func updateSource(tx *sqlx.Tx, newIndex int64, source string) error {
	result, err := tx.Exec("UPDATE source SET index = $1 WHERE name = $2", newIndex, source)
	if err != nil {
		return errors.Wrap(err, "Could not update index")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "Could not read result of source update for %s", source)
	}

	if affected == 0 {
		_, err := tx.Exec("INSERT INTO source (name, index) VALUES ($1, $2)", source, newIndex)
		if err != nil {
			return errors.Wrapf(err, "Could not insert source for %s", source)
		}
	}

	return nil
}

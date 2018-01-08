// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package postgres

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/prizem-io/api/v1"
)

type Ports []api.Port

func (s *Ports) Scan(src interface{}) error {
	asString, ok := src.(string)
	if !ok {
		return error(errors.New("Scan source was not string"))
	}

	parts := strings.Split(asString, "|")
	ports := make([]api.Port, len(parts))
	for i, p := range parts {
		lr := strings.Split(p, ":")
		port, _ := strconv.Atoi(lr[0])
		var secure bool
		if len(lr) > 2 {
			secure, _ = strconv.ParseBool(lr[2])
		}
		ports[i] = api.Port{
			Port:     int32(port),
			Protocol: lr[1],
			Secure:   secure,
		}
	}
	(*s) = ports

	return nil
}

func (s Ports) Value() (driver.Value, error) {
	var buffer bytes.Buffer

	for i, port := range s {
		if i > 0 {
			buffer.WriteString("|")
		}

		buffer.WriteString(fmt.Sprintf("%d:%s:%t", port.Port, port.Protocol, port.Secure))
	}

	return buffer.String(), nil
}

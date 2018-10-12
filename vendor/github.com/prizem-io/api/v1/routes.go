// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

import (
	"time"
)

type (
	// Routes handles retrieving and modifiying services and routes.
	Routes interface {
		GetServices(currentIndex int64) (services []Service, index int64, useCache bool, err error)
		SetService(service Service) (index int64, err error)
		RemoveService(serviceName string) (index int64, err error)
	}

	// Duration defines text marshalling and unmarshalling for `time.Duration`.
	Duration time.Duration

	// RouteInfo encapsulates services and routes as well as global routing rules.
	RouteInfo struct {
		Services []Service     `json:"services"`
		Global   *RoutingRules `json:"global,omitempty"`
	}

	// Service defines a service application and its operations and service-scoped routing rules.
	Service struct {
		Name           string   `json:"name"`
		Namespace      string   `json:"namespace,omitempty"`
		Description    string   `json:"description"`
		Hostname       string   `json:"hostname,omitempty"`
		URIPrefix      string   `json:"uriPrefix"`
		Version        *Version `json:"version,omitempty"`
		Authentication string   `json:"authentication"`
		RoutingRules
		Operations  []Operation  `json:"operations"`
		HealthCheck *HealthCheck `json:"healthCheck,omitempty"`
	}

	// Operation defines a single service operation and its operation-scoped routing rules.
	Operation struct {
		Name       string `json:"name"`
		Method     string `json:"method"`
		URIPattern string `json:"uriPattern"`
		RoutingRules
	}

	// RoutingRules defines the selectors, rewrite rules, retry behavior, and policies for a service or operation.
	RoutingRules struct {
		Selectors    []string        `json:"selectors,omitempty"`
		RewriteRules []Configuration `json:"rewriteRules,omitempty"`
		Timeout      *Duration       `json:"timeout,omitempty"`
		Retry        *Retry          `json:"retry,omitempty"`
		Policies     []Configuration `json:"policies,omitempty"`
		Middleware   interface{}     `json:"-"`
	}

	// Configuration defines loadable parameters for a given middleware type.
	Configuration struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config,omitempty"`
	}

	// Version defines rules for extracting the service/API version from the request.
	Version struct {
		VersionLocations []string `json:"versionLocations,omitempty"`
		DefaultVersion   string   `json:"defaultVersion"`
	}

	// Retry defines retry behavior rule.
	Retry struct {
		Attempts           int      `json:"attempts"`
		ResponseClassifier string   `json:"responseClassifier"`
		PerTryTimeout      Duration `json:"perTryTimeout"`
	}

	// HealthCheck defines the health check behavior.
	HealthCheck struct {
		Timeout            Duration               `json:"timeout"`
		Interval           Duration               `json:"interval"`
		UnhealthyThreshold uint32                 `json:"unhealthyThreshold"`
		HealthyThreshold   uint32                 `json:"healthyThreshold"`
		CheckType          string                 `json:"checkType"`
		CheckConfig        map[string]interface{} `json:"checkConfig,omitempty"`
	}
)

// UnmarshalText converts the JSON bytes into a Duration value.
func (d *Duration) UnmarshalText(b []byte) error {
	// Parse the string to produce a proper time.Time struct.
	pd, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(pd)

	return nil
}

// MarshalText converts a Duration value into JSON bytes.
func (d Duration) MarshalText() ([]byte, error) {
	stringValue := time.Duration(d).String()
	return []byte(stringValue), nil
}

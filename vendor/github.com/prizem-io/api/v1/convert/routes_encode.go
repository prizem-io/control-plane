// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package convert

import (
	"time"

	"github.com/prizem-io/api/v1"
	"github.com/prizem-io/api/v1/proto"
)

func EncodeServices(services []api.Service) []proto.Service {
	if services == nil {
		return nil
	}

	encoded := make([]proto.Service, len(services))
	for i, s := range services {
		encoded[i] = proto.Service{
			Name:           s.Name,
			Namespace:      s.Namespace,
			Description:    s.Description,
			Hostname:       s.Hostname,
			URIPrefix:      s.URIPrefix,
			Version:        EncodeVersion(s.Version),
			Authentication: s.Authentication,
			RoutingRules:   EncodeRoutingRules(s.RoutingRules),
			Operations:     EncodeOperations(s.Operations),
			HealthCheck:    EncodeHealthCheck(s.HealthCheck),
		}
	}
	return encoded
}

func EncodeVersion(version *api.Version) *proto.Version {
	if version == nil {
		return nil
	}

	return &proto.Version{
		VersionLocations: version.VersionLocations,
		DefaultVersion:   version.DefaultVersion,
	}
}

func EncodeOperations(operations []api.Operation) []proto.Operation {
	if operations == nil {
		return nil
	}

	encoded := make([]proto.Operation, len(operations))
	for i, o := range operations {
		encoded[i] = proto.Operation{
			Name:         o.Name,
			Method:       o.Method,
			URIPattern:   o.URIPattern,
			RoutingRules: EncodeRoutingRules(o.RoutingRules),
		}
	}
	return encoded
}

func EncodeRoutingRules(rules api.RoutingRules) proto.RoutingRules {
	return proto.RoutingRules{
		Selectors:    rules.Selectors,
		RewriteRules: EncodeConfigurations(rules.RewriteRules),
		Timeout:      (*time.Duration)(rules.Timeout),
		Retry:        EncodeRetry(rules.Retry),
		Policies:     EncodeConfigurations(rules.Policies),
	}
}

func EncodeRetry(retry *api.Retry) *proto.Retry {
	if retry == nil {
		return nil
	}

	return &proto.Retry{
		Attempts:      int32(retry.Attempts),
		PerTryTimeout: time.Duration(retry.PerTryTimeout),
	}
}

func EncodeConfigurations(configs []api.Configuration) []proto.Configuration {
	if configs == nil {
		return nil
	}

	encoded := make([]proto.Configuration, len(configs))
	for i, c := range configs {
		encoded[i] = proto.Configuration{
			Type:   c.Type,
			Config: EncodeAttributes(c.Config),
		}
	}
	return encoded
}

func EncodeHealthCheck(healthCheck *api.HealthCheck) *proto.HealthCheck {
	if healthCheck == nil {
		return nil
	}

	return &proto.HealthCheck{
		Timeout:            time.Duration(healthCheck.Timeout),
		Interval:           time.Duration(healthCheck.Interval),
		UnhealthyThreshold: healthCheck.UnhealthyThreshold,
		HealthyThreshold:   healthCheck.HealthyThreshold,
		CheckType:          healthCheck.CheckType,
		CheckConfig:        EncodeAttributes(healthCheck.CheckConfig),
	}
}

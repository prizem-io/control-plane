// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package convert

import (
	"github.com/prizem-io/api/v1"
	"github.com/prizem-io/api/v1/proto"
)

func DecodeServices(services []proto.Service) []api.Service {
	if services == nil {
		return nil
	}

	decoded := make([]api.Service, len(services))
	for i, s := range services {
		decoded[i] = api.Service{
			Name:           s.Name,
			Namespace:      s.Namespace,
			Description:    s.Description,
			Hostname:       s.Hostname,
			URIPrefix:      s.URIPrefix,
			Version:        DecodeVersion(s.Version),
			Authentication: s.Authentication,
			RoutingRules:   DecodeRoutingRules(s.RoutingRules),
			Operations:     DecodeOperations(s.Operations),
			HealthCheck:    DecodeHealthCheck(s.HealthCheck),
		}
	}
	return decoded
}

func DecodeVersion(version *proto.Version) *api.Version {
	if version == nil {
		return nil
	}

	return &api.Version{
		VersionLocations: version.VersionLocations,
		DefaultVersion:   version.DefaultVersion,
	}
}

func DecodeOperations(operations []proto.Operation) []api.Operation {
	if operations == nil {
		return nil
	}

	decoded := make([]api.Operation, len(operations))
	for i, o := range operations {
		decoded[i] = api.Operation{
			Name:         o.Name,
			Method:       o.Method,
			URIPattern:   o.URIPattern,
			RoutingRules: DecodeRoutingRules(o.RoutingRules),
		}
	}
	return decoded
}

func DecodeRoutingRules(rules proto.RoutingRules) api.RoutingRules {
	return api.RoutingRules{
		Selectors:    rules.Selectors,
		RewriteRules: DecodeConfigurations(rules.RewriteRules),
		Timeout:      (*api.Duration)(rules.Timeout),
		Retry:        DecodeRetry(rules.Retry),
		Policies:     DecodeConfigurations(rules.Policies),
	}
}

func DecodeRetry(retry *proto.Retry) *api.Retry {
	if retry == nil {
		return nil
	}

	return &api.Retry{
		Attempts:      int(retry.Attempts),
		PerTryTimeout: api.Duration(retry.PerTryTimeout),
	}
}

func DecodeConfigurations(configs []proto.Configuration) []api.Configuration {
	if configs == nil {
		return nil
	}

	decoded := make([]api.Configuration, len(configs))
	for i, c := range configs {
		decoded[i] = api.Configuration{
			Type:   c.Type,
			Config: DecodeAttributes(c.Config),
		}
	}
	return decoded
}

func DecodeHealthCheck(healthCheck *proto.HealthCheck) *api.HealthCheck {
	if healthCheck == nil {
		return nil
	}

	return &api.HealthCheck{
		Timeout:            api.Duration(healthCheck.Timeout),
		Interval:           api.Duration(healthCheck.Interval),
		UnhealthyThreshold: healthCheck.UnhealthyThreshold,
		HealthyThreshold:   healthCheck.HealthyThreshold,
		CheckType:          healthCheck.CheckType,
		CheckConfig:        DecodeAttributes(healthCheck.CheckConfig),
	}
}

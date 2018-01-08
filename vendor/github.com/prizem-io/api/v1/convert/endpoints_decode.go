// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package convert

import (
	"net"

	"github.com/satori/go.uuid"

	"github.com/prizem-io/api/v1"
	"github.com/prizem-io/api/v1/proto"
)

// DecodeNodes decodes `[]proto.Node` to `[]api.Node`.
func DecodeNodes(nodes []proto.Node) []api.Node {
	if nodes == nil {
		return nil
	}

	pnodes := make([]api.Node, len(nodes))
	for i, n := range nodes {
		pnodes[i] = DecodeNode(n)
	}
	return pnodes
}

// DecodeNode decodes `proto.Node` to `api.Node`.
func DecodeNode(in proto.Node) api.Node {
	uuid.NewV4()
	return api.Node{
		ID:         uuid.FromStringOrNil(in.ID),
		Geography:  in.Geography,
		Datacenter: in.Datacenter,
		Address:    net.ParseIP(in.Address),
		Metadata:   DecodeAttributes(in.Metadata),
		Services:   DecodeServiceInstances(in.Services),
	}
}

// DecodeServiceInstances decodes `[]proto.ServiceInstance` to `[]api.ServiceInstance`.
func DecodeServiceInstances(services []proto.ServiceInstance) []api.ServiceInstance {
	if services == nil {
		return nil
	}

	decoded := make([]api.ServiceInstance, len(services))
	for i, si := range services {
		decoded[i] = api.ServiceInstance{
			ID:        uuid.FromStringOrNil(si.ID),
			Service:   si.Service,
			Name:      si.Name,
			Namespace: si.Namespace,
			Principal: si.Principal,
			Owner:     si.Owner,
			Container: DecodeContainer(si.Container),
			Ports:     DecodePorts(si.Ports),
			Metadata:  DecodeAttributes(si.Metadata),
			Labels:    si.Labels,
		}
	}
	return decoded
}

// DecodeContainer decodes `*proto.Container` to `*api.Container`.
func DecodeContainer(container *proto.Container) *api.Container {
	if container == nil {
		return nil
	}

	return &api.Container{
		Name:  container.Name,
		Image: container.Image,
	}
}

// DecodePorts decodes `[]proto.Port` to `[]api.Port`.
func DecodePorts(ports []proto.Port) []api.Port {
	if ports == nil {
		return nil
	}

	decoded := make([]api.Port, len(ports))
	for i, si := range ports {
		decoded[i] = api.Port{
			Port:     si.Port,
			Protocol: si.Protocol,
			Secure:   si.Secure,
		}
	}
	return decoded
}

// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package convert

import (
	"github.com/prizem-io/api/v1"
	"github.com/prizem-io/api/v1/proto"
)

func EncodeNodes(nodes []api.Node) []proto.Node {
	if nodes == nil {
		return nil
	}

	pnodes := make([]proto.Node, len(nodes))
	for i, n := range nodes {
		pnodes[i] = EncodeNode(n)
	}
	return pnodes
}

func EncodeNode(in api.Node) proto.Node {
	return proto.Node{
		ID:         in.ID.String(),
		Geography:  in.Geography,
		Datacenter: in.Datacenter,
		Address:    in.Address.String(),
		Metadata:   EncodeAttributes(in.Metadata),
		Services:   EncodeServiceInstances(in.Services),
	}
}

func EncodeServiceInstances(services []api.ServiceInstance) []proto.ServiceInstance {
	if services == nil {
		return nil
	}

	encoded := make([]proto.ServiceInstance, len(services))
	for i, si := range services {
		encoded[i] = proto.ServiceInstance{
			ID:        si.ID.String(),
			Service:   si.Service,
			Name:      si.Name,
			Namespace: si.Namespace,
			Principal: si.Principal,
			Owner:     si.Owner,
			Container: EncodeContainer(si.Container),
			Ports:     EncodePorts(si.Ports),
			Metadata:  EncodeAttributes(si.Metadata),
			Labels:    si.Labels,
		}
	}
	return encoded
}

func EncodeContainer(container *api.Container) *proto.Container {
	if container == nil {
		return nil
	}

	return &proto.Container{
		Name:  container.Name,
		Image: container.Image,
	}
}

func EncodePorts(ports []api.Port) []proto.Port {
	if ports == nil {
		return nil
	}

	encoded := make([]proto.Port, len(ports))
	for i, si := range ports {
		encoded[i] = proto.Port{
			Port:     si.Port,
			Protocol: si.Protocol,
			Secure:   si.Secure,
		}
	}
	return encoded
}

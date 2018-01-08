// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package api

import (
	"net"

	uuid "github.com/satori/go.uuid"
)

type (
	// Endpoints handles retrieving and modifiying endpoints (nodes and service instances).
	Endpoints interface {
		GetEndpoints(currentIndex int64) (nodes []Node, index int64, useCache bool, err error)
		AddEndpoints(node Node) (modification *Modification, err error)
		RemoveEndpoints(nodeID string, serviceNames ...string) (modification *Modification, err error)
	}

	// Modification defines the post-modification state after adding or removing nodes and service instances.
	Modification struct {
		Index    int64    `json:"index"`
		Node     bool     `json:"nodeAffected"`
		Services []string `json:"services"`
	}

	// Metadata is user-defined key/value pairs that can be applied to nodes and services instances.
	Metadata map[string]interface{}

	// EndpointInfo encapsulates node and service instance information in an object payload.
	EndpointInfo struct {
		Nodes []Node `json:"nodes"`
	}

	// Node represents a single running server node or sidecar proxy.
	Node struct {
		ID         uuid.UUID         `json:"id"`
		Geography  string            `json:"geography"`
		Datacenter string            `json:"datacenter"`
		Address    net.IP            `json:"address"`
		Metadata   Metadata          `json:"metadata,omitempty"`
		Services   []ServiceInstance `json:"services,omitempty"`
	}

	// Ports provides convenience methods on top if `[]Port`
	Ports []Port

	// ServiceInstance defines a single running instance of a service.
	ServiceInstance struct {
		ID        uuid.UUID           `json:"id"`
		Service   string              `json:"service"`
		Name      string              `json:"name,omitempty"`
		Namespace string              `json:"namespace,omitempty"`
		Principal string              `json:"principal,omitempty"`
		Owner     string              `json:"owner,omitempty"`
		Container *Container          `json:"container,omitempty"`
		Ports     Ports               `json:"ports"`
		Metadata  Metadata            `json:"metadata,omitempty"`
		Labels    []string            `json:"labels,omitempty"`
		LabelSet  map[string]struct{} `json:"-"`
	}

	// Container encapsulates optional container information.
	Container struct {
		Name  string `json:"name,omitempty"`
		Image string `json:"image,omitempty"`
	}

	// Port defines the port and protcol a service is listening on.
	Port struct {
		Port     int32  `json:"port"`
		Protocol string `json:"protocol"`
		Secure   bool   `json:"secure"`
	}
)

// Find returns the first occurrence of a port that matches one of the desired protocols (in preference order).
func (p Ports) Find(protocols ...string) (Port, bool) {
	for _, protocol := range protocols {
		for _, port := range p {
			if port.Protocol == protocol {
				return port, true
			}
		}
	}

	return Port{}, false
}

// Prepare initializes optimized data structures after loading or unmarshalling.
func (e *EndpointInfo) Prepare() {
	for i := range e.Nodes {
		e.Nodes[i].Prepare()
	}
}

// Prepare initializes optimized data structures after loading or unmarshalling.
func (n *Node) Prepare() {
	for i := range n.Services {
		n.Services[i].Prepare()
	}
}

// Prepare initializes optimized data structures after loading or unmarshalling.
func (s *ServiceInstance) Prepare() {
	s.LabelSet = make(map[string]struct{}, len(s.LabelSet))
	for _, label := range s.Labels {
		s.LabelSet[label] = struct{}{}
	}
}

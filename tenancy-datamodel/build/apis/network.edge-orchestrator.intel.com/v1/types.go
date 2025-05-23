// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by nexus. DO NOT EDIT.

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/common"
)

// +k8s:openapi-gen=true
type Child struct {
	Group string `json:"group" yaml:"group"`
	Kind  string `json:"kind" yaml:"kind"`
	Name  string `json:"name" yaml:"name"`
}

// +k8s:openapi-gen=true
type Link struct {
	Group string `json:"group" yaml:"group"`
	Kind  string `json:"kind" yaml:"kind"`
	Name  string `json:"name" yaml:"name"`
}

// +k8s:openapi-gen=true
type SyncerStatus struct {
	EtcdVersion    int64 `json:"etcdVersion, omitempty" yaml:"etcdVersion, omitempty"`
	CRGenerationId int64 `json:"cRGenerationId, omitempty" yaml:"cRGenerationId, omitempty"`
}

// +k8s:openapi-gen=true
type NexusStatus struct {
	SourceGeneration int64        `json:"sourceGeneration, omitempty" yaml:"sourceGeneration, omitempty"`
	RemoteGeneration int64        `json:"remoteGeneration, omitempty" yaml:"remoteGeneration, omitempty"`
	SyncerStatus     SyncerStatus `json:"syncerStatus, omitempty" yaml:"syncerStatus, omitempty"`
}

/* ------------------- CRDs definitions ------------------- */

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type Network struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec              NetworkSpec        `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status            NetworkNexusStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// +k8s:openapi-gen=true
type NetworkNexusStatus struct {
	Status NetworkStatus `json:"status,omitempty" yaml:"status,omitempty"`
	Nexus  NexusStatus   `json:"nexus,omitempty" yaml:"nexus,omitempty"`
}

func (c *Network) CRDName() string {
	return "networks.network.edge-orchestrator.intel.com"
}

func (c *Network) DisplayName() string {
	if c.GetLabels() != nil {
		return c.GetLabels()[common.DisplayNameLabel]
	}
	return ""
}

// +k8s:openapi-gen=true
type NetworkSpec struct {
	Type        NetworkType `json:"type" yaml:"type"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NetworkList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata" yaml:"metadata"`
	Items           []Network `json:"items" yaml:"items"`
}

// +k8s:openapi-gen=true
type NetworkStatus struct {
	CurrentState string `json:"currentState" yaml:"currentState"`
}

//nolint:revive // Per requirement.
type NetworkType string

const AppplicationMesh NetworkType = "application-mesh"

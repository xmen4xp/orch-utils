// Copyright The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "/build/client/clientset/versioned/typed/policypkg.tsm.tanzu.vmware.com/v1"

	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakePolicypkgTsmV1 struct {
	*testing.Fake
}

func (c *FakePolicypkgTsmV1) ACPConfigs() v1.ACPConfigInterface {
	return &FakeACPConfigs{c}
}

func (c *FakePolicypkgTsmV1) AccessControlPolicies() v1.AccessControlPolicyInterface {
	return &FakeAccessControlPolicies{c}
}

func (c *FakePolicypkgTsmV1) VMpolicies() v1.VMpolicyInterface {
	return &FakeVMpolicies{c}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakePolicypkgTsmV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}

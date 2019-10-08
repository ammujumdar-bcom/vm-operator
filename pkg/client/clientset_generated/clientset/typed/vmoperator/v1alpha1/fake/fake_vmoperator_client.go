/* **********************************************************
 * Copyright 2019 VMware, Inc.  All rights reserved. -- VMware Confidential
 * **********************************************************/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/vmware-tanzu/vm-operator/pkg/client/clientset_generated/clientset/typed/vmoperator/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeVmoperatorV1alpha1 struct {
	*testing.Fake
}

func (c *FakeVmoperatorV1alpha1) VirtualMachines(namespace string) v1alpha1.VirtualMachineInterface {
	return &FakeVirtualMachines{c, namespace}
}

func (c *FakeVmoperatorV1alpha1) VirtualMachineClasses() v1alpha1.VirtualMachineClassInterface {
	return &FakeVirtualMachineClasses{c}
}

func (c *FakeVmoperatorV1alpha1) VirtualMachineImages(namespace string) v1alpha1.VirtualMachineImageInterface {
	return &FakeVirtualMachineImages{c, namespace}
}

func (c *FakeVmoperatorV1alpha1) VirtualMachineServices(namespace string) v1alpha1.VirtualMachineServiceInterface {
	return &FakeVirtualMachineServices{c, namespace}
}

func (c *FakeVmoperatorV1alpha1) VirtualMachineSetResourcePolicies(namespace string) v1alpha1.VirtualMachineSetResourcePolicyInterface {
	return &FakeVirtualMachineSetResourcePolicies{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeVmoperatorV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}

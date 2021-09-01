// Copyright (c) 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	"context"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"

	"github.com/vmware-tanzu/vm-operator/pkg/vmprovider"
)

// This Fake Provider is supposed to simulate an actual VM provider.
// That includes maintaining states for objects created through this
// provider, and query/get APIs reflecting those states. As a convenience
// to simulate scenarios, this provider also exposes a few function variables
// to override certain behaviors. The functionality of this provider is
// expected to evolve as more tests get added in the future.

type funcs struct {
	DoesVirtualMachineExistFn         func(ctx context.Context, vm *v1alpha1.VirtualMachine) (bool, error)
	CreateVirtualMachineFn            func(ctx context.Context, vm *v1alpha1.VirtualMachine, vmConfigArgs vmprovider.VMConfigArgs) error
	UpdateVirtualMachineFn            func(ctx context.Context, vm *v1alpha1.VirtualMachine, vmConfigArgs vmprovider.VMConfigArgs) error
	DeleteVirtualMachineFn            func(ctx context.Context, vm *v1alpha1.VirtualMachine) error
	GetVirtualMachineGuestHeartbeatFn func(ctx context.Context, vm *v1alpha1.VirtualMachine) (v1alpha1.GuestHeartbeatStatus, error)

	ListVirtualMachineImagesFromContentLibraryFn func(ctx context.Context, cl v1alpha1.ContentLibraryProvider, currentCLImages map[string]v1alpha1.VirtualMachineImage) ([]*v1alpha1.VirtualMachineImage, error)
	DoesContentLibraryExistFn                    func(ctx context.Context, cl *v1alpha1.ContentLibraryProvider) (bool, error)

	UpdateVcPNIDFn                  func(ctx context.Context, vcPNID, vcPort string) error
	ClearSessionsAndClientFn        func(ctx context.Context)
	DeleteNamespaceSessionInCacheFn func(ctx context.Context, namespace string)

	CreateOrUpdateVirtualMachineSetResourcePolicyFn func(ctx context.Context, rp *v1alpha1.VirtualMachineSetResourcePolicy) error
	IsVirtualMachineSetResourcePolicyReadyFn        func(ctx context.Context, azName string, rp *v1alpha1.VirtualMachineSetResourcePolicy) (bool, error)
	DeleteVirtualMachineSetResourcePolicyFn         func(ctx context.Context, rp *v1alpha1.VirtualMachineSetResourcePolicy) error
	ComputeClusterCPUMinFrequencyFn                 func(ctx context.Context) error
}

type VMProvider struct {
	sync.Mutex
	funcs
	vmMap             map[client.ObjectKey]*v1alpha1.VirtualMachine
	resourcePolicyMap map[client.ObjectKey]*v1alpha1.VirtualMachineSetResourcePolicy
}

var _ vmprovider.VirtualMachineProviderInterface = &VMProvider{}

func (s *VMProvider) Reset() {
	s.Lock()
	defer s.Unlock()

	s.funcs = funcs{}
	s.vmMap = make(map[client.ObjectKey]*v1alpha1.VirtualMachine)
	s.resourcePolicyMap = make(map[client.ObjectKey]*v1alpha1.VirtualMachineSetResourcePolicy)
}

func (s *VMProvider) DoesVirtualMachineExist(ctx context.Context, vm *v1alpha1.VirtualMachine) (bool, error) {
	s.Lock()
	defer s.Unlock()
	if s.DoesVirtualMachineExistFn != nil {
		return s.DoesVirtualMachineExistFn(ctx, vm)
	}
	objectKey := client.ObjectKey{
		Namespace: vm.Namespace,
		Name:      vm.Name,
	}
	_, ok := s.vmMap[objectKey]
	return ok, nil
}

func (s *VMProvider) CreateVirtualMachine(ctx context.Context, vm *v1alpha1.VirtualMachine, vmConfigArgs vmprovider.VMConfigArgs) error {
	s.Lock()
	defer s.Unlock()
	if s.CreateVirtualMachineFn != nil {
		return s.CreateVirtualMachineFn(ctx, vm, vmConfigArgs)
	}
	s.addToVMMap(vm)
	vm.Status.Phase = v1alpha1.Created
	return nil
}

func (s *VMProvider) UpdateVirtualMachine(ctx context.Context, vm *v1alpha1.VirtualMachine, vmConfigArgs vmprovider.VMConfigArgs) error {
	s.Lock()
	defer s.Unlock()
	if s.UpdateVirtualMachineFn != nil {
		return s.UpdateVirtualMachineFn(ctx, vm, vmConfigArgs)
	}
	s.addToVMMap(vm)
	return nil
}

func (s *VMProvider) DeleteVirtualMachine(ctx context.Context, vm *v1alpha1.VirtualMachine) error {
	s.Lock()
	defer s.Unlock()
	if s.DeleteVirtualMachineFn != nil {
		return s.DeleteVirtualMachineFn(ctx, vm)
	}
	s.deleteFromVMMap(vm)
	return nil
}

func (s *VMProvider) GetVirtualMachineGuestHeartbeat(ctx context.Context, vm *v1alpha1.VirtualMachine) (v1alpha1.GuestHeartbeatStatus, error) {
	s.Lock()
	defer s.Unlock()
	if s.GetVirtualMachineGuestHeartbeatFn != nil {
		return s.GetVirtualMachineGuestHeartbeatFn(ctx, vm)
	}
	return "", nil
}

func (s *VMProvider) Initialize(stop <-chan struct{}) {}

func (s *VMProvider) Name() string {
	return "fake"
}

func (s *VMProvider) GetClusterID(ctx context.Context, namespace string) (string, error) {
	return "domain-c8", nil
}

func (s *VMProvider) CreateOrUpdateVirtualMachineSetResourcePolicy(ctx context.Context, resourcePolicy *v1alpha1.VirtualMachineSetResourcePolicy) error {
	s.Lock()
	defer s.Unlock()

	if s.CreateOrUpdateVirtualMachineSetResourcePolicyFn != nil {
		return s.CreateOrUpdateVirtualMachineSetResourcePolicyFn(ctx, resourcePolicy)
	}
	s.addToResourcePolicyMap(resourcePolicy)

	return nil
}

func (s *VMProvider) IsVirtualMachineSetResourcePolicyReady(ctx context.Context, azName string, resourcePolicy *v1alpha1.VirtualMachineSetResourcePolicy) (bool, error) {
	s.Lock()
	defer s.Unlock()

	if s.IsVirtualMachineSetResourcePolicyReadyFn != nil {
		return s.IsVirtualMachineSetResourcePolicyReadyFn(ctx, azName, resourcePolicy)
	}
	objectKey := client.ObjectKey{
		Namespace: resourcePolicy.Namespace,
		Name:      resourcePolicy.Name,
	}
	_, found := s.resourcePolicyMap[objectKey]

	return found, nil
}

func (s *VMProvider) DeleteVirtualMachineSetResourcePolicy(ctx context.Context, resourcePolicy *v1alpha1.VirtualMachineSetResourcePolicy) error {
	s.Lock()
	defer s.Unlock()

	if s.DeleteVirtualMachineSetResourcePolicyFn != nil {
		return s.DeleteVirtualMachineSetResourcePolicyFn(ctx, resourcePolicy)
	}
	s.deleteFromResourcePolicyMap(resourcePolicy)

	return nil
}

func (s *VMProvider) ComputeClusterCPUMinFrequency(ctx context.Context) error {
	s.Lock()
	defer s.Unlock()
	if s.ComputeClusterCPUMinFrequencyFn != nil {
		return s.ComputeClusterCPUMinFrequencyFn(ctx)
	}

	return nil
}

func (s *VMProvider) UpdateVcPNID(ctx context.Context, vcPNID, vcPort string) error {
	s.Lock()
	defer s.Unlock()
	if s.UpdateVcPNIDFn != nil {
		return s.UpdateVcPNIDFn(ctx, vcPNID, vcPort)
	}
	return nil
}

func (s *VMProvider) ClearSessionsAndClient(ctx context.Context) {
	s.Lock()
	defer s.Unlock()

	if s.ClearSessionsAndClientFn != nil {
		s.ClearSessionsAndClientFn(ctx)
	}
}

func (s *VMProvider) DeleteNamespaceSessionInCache(ctx context.Context, namespace string) error {
	s.Lock()
	defer s.Unlock()

	if s.DeleteNamespaceSessionInCacheFn != nil {
		s.DeleteNamespaceSessionInCacheFn(ctx, namespace)
	}

	return nil
}

func (s *VMProvider) DoesContentLibraryExist(ctx context.Context, contentLibrary *v1alpha1.ContentLibraryProvider) (bool, error) {
	s.Lock()
	defer s.Unlock()
	if s.DoesContentLibraryExistFn != nil {
		return s.DoesContentLibraryExistFn(ctx, contentLibrary)
	}

	return true, nil
}

func (s *VMProvider) ListVirtualMachineImagesFromContentLibrary(ctx context.Context, cl v1alpha1.ContentLibraryProvider, currentCLImages map[string]v1alpha1.VirtualMachineImage) ([]*v1alpha1.VirtualMachineImage, error) {
	s.Lock()
	defer s.Unlock()

	if s.ListVirtualMachineImagesFromContentLibraryFn != nil {
		return s.ListVirtualMachineImagesFromContentLibraryFn(ctx, cl, currentCLImages)
	}

	// No-op for now.
	return []*v1alpha1.VirtualMachineImage{}, nil
}

func (s *VMProvider) ListVirtualMachineImages(ctx context.Context, namespace string) ([]*v1alpha1.VirtualMachineImage, error) {
	return []*v1alpha1.VirtualMachineImage{}, nil
}

func (s *VMProvider) GetVirtualMachineImage(ctx context.Context, namespace, name string) (*v1alpha1.VirtualMachineImage, error) {
	return nil, nil
}

func (s *VMProvider) addToVMMap(vm *v1alpha1.VirtualMachine) {
	objectKey := client.ObjectKey{
		Namespace: vm.Namespace,
		Name:      vm.Name,
	}
	s.vmMap[objectKey] = vm
}

func (s *VMProvider) deleteFromVMMap(vm *v1alpha1.VirtualMachine) {
	objectKey := client.ObjectKey{
		Namespace: vm.Namespace,
		Name:      vm.Name,
	}
	delete(s.vmMap, objectKey)
}

func (s *VMProvider) addToResourcePolicyMap(rp *v1alpha1.VirtualMachineSetResourcePolicy) {
	objectKey := client.ObjectKey{
		Namespace: rp.Namespace,
		Name:      rp.Name,
	}

	s.resourcePolicyMap[objectKey] = rp
}

func (s *VMProvider) deleteFromResourcePolicyMap(rp *v1alpha1.VirtualMachineSetResourcePolicy) {
	objectKey := client.ObjectKey{
		Namespace: rp.Namespace,
		Name:      rp.Name,
	}
	delete(s.resourcePolicyMap, objectKey)
}

func NewVMProvider() *VMProvider {
	provider := VMProvider{
		vmMap:             map[client.ObjectKey]*v1alpha1.VirtualMachine{},
		resourcePolicyMap: map[client.ObjectKey]*v1alpha1.VirtualMachineSetResourcePolicy{},
	}
	return &provider
}

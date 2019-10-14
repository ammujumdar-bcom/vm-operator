/* **********************************************************
 * Copyright 2018-2019 VMware, Inc.  All rights reserved. -- VMware Confidential
 * **********************************************************/

package vsphere

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"strings"

	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/vcenter"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware-tanzu/vm-operator/pkg/vmprovider"

	"github.com/vmware-tanzu/vm-operator/pkg/apis/vmoperator/v1alpha1"
	ncpv1alpha1 "github.com/vmware-tanzu/vm-operator/external/ncp/api/v1alpha1"
	clientset "gitlab.eng.vmware.com/guest-clusters/ncp-client/pkg/client/clientset/versioned"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	vimTypes "github.com/vmware/govmomi/vim25/types"
	res "github.com/vmware-tanzu/vm-operator/pkg/vmprovider/providers/vsphere/resources"
)

type Session struct {
	client *Client

	ncpClient clientset.Interface

	Finder       *find.Finder
	datacenter   *object.Datacenter
	cluster      *object.ClusterComputeResource
	folder       *object.Folder
	resourcepool *object.ResourcePool
	datastore    *object.Datastore
	network      object.NetworkReference
	contentlib   *library.Library
	creds        *VSphereVmProviderCredentials
	extraConfig  map[string]string
}

func NewSessionAndConfigure(ctx context.Context, config *VSphereVmProviderConfig, ncpclient clientset.Interface) (*Session, error) {
	c, err := NewClient(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client for new session")
	}

	s := &Session{
		client:    c,
		ncpClient: ncpclient,
	}

	if err = s.initSession(ctx, config); err != nil {
		s.Logout(ctx)
		return nil, err
	}

	if err = s.ConfigureContent(ctx, config.ContentSource); err != nil {
		return nil, err
	}

	log.Info("New session created and configured", "Session", s.String())
	return s, nil
}

func (s *Session) initSession(ctx context.Context, config *VSphereVmProviderConfig) error {
	s.Finder = find.NewFinder(s.client.VimClient(), false)

	dc, err := s.Finder.Datacenter(ctx, config.Datacenter)
	if err != nil {
		return errors.Wrapf(err, "failed to init Datacenter %q", config.Datacenter)
	}

	s.datacenter = dc
	s.Finder.SetDatacenter(dc)

	// not necessary for vmimage list/get from Content Library
	if config.ResourcePool != "" {
		s.resourcepool, err = GetResourcePool(ctx, s.Finder, config.ResourcePool)
		if err != nil {
			return errors.Wrapf(err, "failed to init Resource Pool %q", config.ResourcePool)
		}
	}

	// not necessary for vmimage list/get from Content Library
	if config.Folder != "" {
		s.folder, err = GetVMFolder(ctx, s.Finder, config.Folder)
		if err != nil {
			return errors.Wrapf(err, "failed to init folder %q", config.Folder)
		}
	}

	// not necessary for vmimage list/get from Content Library
	if s.resourcepool != nil {
		s.cluster, err = GetResourcePoolOwner(ctx, s.resourcepool)
		if err != nil {
			return errors.Wrapf(err, "failed to init cluster %q", config.ResourcePool)
		}
	}

	// not necessary for vmimage list/get from Content Library
	if config.Datastore != "" {
		s.datastore, err = s.Finder.Datastore(ctx, config.Datastore)
		if err != nil {
			return errors.Wrapf(err, "failed to init Datastore %q", config.Datastore)
		}
	}

	// Network setting is optional
	if config.Network != "" {
		s.network, err = s.Finder.Network(ctx, config.Network)
		if err != nil {
			return errors.Wrapf(err, "failed to init Network %q", config.Network)
		}
		log.Info("Using default network", "network", config.Network)
	}

	// Allow for the option to specify extraConfig to be applied to all VMs
	if jsonExtraConfig := os.Getenv("JSON_EXTRA_CONFIG"); jsonExtraConfig != "" {
		s.extraConfig = make(map[string]string)
		if err := json.Unmarshal([]byte(jsonExtraConfig), &s.extraConfig); err != nil {
			return errors.Wrapf(err, "Unable to parse Json ExtraConfig")
		}
		log.Info("Using Json extraConfig", "extraConfig", s.extraConfig)
	}

	s.creds = config.VcCreds

	return nil
}

func (s *Session) ConfigureContent(ctx context.Context, contentSource string) error {
	if contentSource == "" {
		s.contentlib = nil
		return nil
	}

	var err error
	if err = s.WithRestClient(ctx, func(c *rest.Client) error {
		libManager := library.NewManager(c)
		s.contentlib, err = libManager.GetLibraryByID(ctx, contentSource)
		// TODO: Below code needs to be removed before 1.0. Allowing a name to be specified for test environments and prevent any breakages.
		if err != nil {
			log.Error(err, "GetLibrarybyID failed: Trying GetLibrarybyName")
			s.contentlib, err = libManager.GetLibraryByName(ctx, contentSource)
		}
		return err
	}); err != nil {
		return errors.Wrapf(err, "failed to init Content Library %q", contentSource)
	}

	return nil
}

func (s *Session) Logout(ctx context.Context) {
	s.client.Logout(ctx)
}

func (s *Session) ListVirtualMachineImagesFromCL(ctx context.Context, namespace string) ([]*v1alpha1.VirtualMachineImage, error) {
	var items []library.Item
	var err error
	err = s.WithRestClient(ctx, func(c *rest.Client) error {
		items, err = library.NewManager(c).GetLibraryItems(ctx, s.contentlib.ID)
		return err
	})
	if err != nil {
		return nil, err
	}

	var images []*v1alpha1.VirtualMachineImage
	for _, item := range items {
		if IsSupportedDeployType(item.Type) {
			var vmOpts OvfPropertyRetriever = vmOptions{}
			virtualMachineImage, err := LibItemToVirtualMachineImage(ctx, s, &item, namespace, DoNotAnnotateVmImage, vmOpts)
			if err != nil {
				return nil, err
			}
			images = append(images, virtualMachineImage)
		}
	}

	return images, err
}

func (s *Session) GetVirtualMachineImageFromCL(ctx context.Context, name string, namespace string) (*v1alpha1.VirtualMachineImage, error) {
	var item *library.Item

	err := s.WithRestClient(ctx, func(c *rest.Client) error {
		itemIDs, err := library.NewManager(c).FindLibraryItems(ctx, library.FindItem{LibraryID: s.contentlib.ID, Name: name})
		if err != nil {
			return err
		}

		if len(itemIDs) > 0 {
			//Handle multiple IDs found as an error or return the first one?
			item, err = library.NewManager(c).GetLibraryItem(ctx, itemIDs[0])
		}
		return err
	})

	if err != nil {
		return nil, err
	}
	//Return nil when the image with 'name' is not found in CL
	if item == nil {
		return nil, errors.Errorf("item: %v is not found in CL", name)
	}
	//if not a supported type return nil
	if !IsSupportedDeployType(item.Type) {
		return nil, errors.Errorf("item: %v not a supported type", item.Name)
	}

	var vmOpts OvfPropertyRetriever = vmOptions{}

	virtualMachineImage, err := LibItemToVirtualMachineImage(ctx, s, item, namespace, AnnotateVmImage, vmOpts)

	if err != nil {
		return nil, err
	}

	return virtualMachineImage, nil
}

func (s *Session) ListVirtualMachines(ctx context.Context, path string) ([]*res.VirtualMachine, error) {
	var vms []*res.VirtualMachine

	objVms, err := s.Finder.VirtualMachineList(ctx, path)
	if err != nil {
		switch err.(type) {
		case *find.NotFoundError, *find.DefaultNotFoundError:
			return vms, nil
		default:
			return nil, err
		}
	}

	for _, objVm := range objVms {
		if resVm, err := res.NewVMFromObject(objVm); err == nil {
			vms = append(vms, resVm)
		}
	}

	return vms, nil
}

func (s *Session) GetVirtualMachine(ctx context.Context, name string) (*res.VirtualMachine, error) {
	return s.lookupVm(ctx, name)
}

func (s *Session) CreateVirtualMachine(ctx context.Context, vm *v1alpha1.VirtualMachine,
	vmClass v1alpha1.VirtualMachineClass, vmMetadata vmprovider.VirtualMachineMetadata) (*res.VirtualMachine, error) {
	deviceSpecs, err := s.deviceSpecsFromVM(ctx, vm)
	if err != nil {
		return nil, err
	}

	name := vm.Name
	configSpec, err := s.configSpecFromClassSpec(name, &vm.Spec, &vmClass.Spec, vmMetadata, deviceSpecs)
	if err != nil {
		return nil, err
	}

	resVm, err := s.createVm(ctx, name, configSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new VM %q", name)
	}

	return resVm, nil
}

func (s *Session) CloneVirtualMachine(ctx context.Context, vm *v1alpha1.VirtualMachine,
	vmClass v1alpha1.VirtualMachineClass, vmMetadata vmprovider.VirtualMachineMetadata, profileID string) (*res.VirtualMachine, error) {
	name := vm.Name

	if s.contentlib != nil {
		image, err := s.GetVirtualMachineImageFromCL(ctx, vm.Spec.ImageName, vm.Namespace)
		if err != nil {
			return nil, err
		}

		log.Info("Going to deploy ovf", "imageName", image.ObjectMeta.Name, "vmName", name, "profileID", profileID)
		deployedVm, err := s.deployOvf(ctx, image.Status.Uuid, name, profileID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to deploy new VM %q from %q", name, vm.Spec.ImageName)
		}
		// Create network resource and reconfigure VM
		deviceSpecs, err := s.deviceChangeSpecs(ctx, vm, deployedVm)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate device change spec for VM %q", name)
		}
		// configure VM device
		err = deployedVm.Reconfigure(ctx, &vimTypes.VirtualMachineConfigSpec{
			DeviceChange: deviceSpecs,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to reconfigure VM %q", name)
		}

		return deployedVm, nil
	}

	resSrcVm, err := s.lookupVm(ctx, vm.Spec.ImageName)
	if err != nil {
		return nil, err
	}

	cloneSpec, err := s.getCloneSpec(ctx, name, resSrcVm, vm, &vmClass.Spec, vmMetadata, profileID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create clone spec from %q", resSrcVm.Name)
	}

	cloneResVm, err := s.cloneVm(ctx, resSrcVm, cloneSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to clone new VM %q from %q", name, resSrcVm.Name)
	}

	return cloneResVm, nil
}

func (s *Session) DeleteVirtualMachine(ctx context.Context, vm *v1alpha1.VirtualMachine) error {
	resVm, err := s.lookupVm(ctx, vm.Name)
	if err != nil {
		return err
	}

	err = resVm.Delete(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to delete VM %q", vm.Name)
	}

	return nil
}

func (s *Session) lookupVm(ctx context.Context, name string) (*res.VirtualMachine, error) {
	objVm, err := s.Finder.VirtualMachine(ctx, name)
	if err != nil {
		return nil, err
	}

	return res.NewVMFromObject(objVm)
}

func memoryQuantityToMb(q resource.Quantity) int64 {
	return int64(math.Ceil(float64(q.Value()) / float64(1024*1024)))
}

func cpuQuantityToMhz(q resource.Quantity) int64 {
	return int64(math.Ceil(float64(q.Value()) / float64(1000*1000)))
}

func (s *Session) deviceSpecsFromVM(ctx context.Context, vm *v1alpha1.VirtualMachine) ([]vimTypes.BaseVirtualDeviceConfigSpec, error) {
	var deviceSpecs []vimTypes.BaseVirtualDeviceConfigSpec

	// The clients should ensure that existing device keys are not reused as temporary key values for the new device to be added, hence
	//  use unique negative integers as temporary keys.
	key := int32(-100)
	for _, vif := range vm.Spec.NetworkInterfaces {
		np, err := s.getNetworkProviderByType(vif.NetworkType)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get network provider")
		}
		dev, err := np.CreateVnic(ctx, vm, &vif)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create vnic '%v'", vif)
		}
		dev = setVnicKey(dev, key)
		deviceSpecs = append(deviceSpecs, &vimTypes.VirtualDeviceConfigSpec{
			Device:    dev,
			Operation: vimTypes.VirtualDeviceConfigSpecOperationAdd,
		})

		key--
	}
	return deviceSpecs, nil
}

func (s *Session) deviceChangeSpecs(ctx context.Context, vm *v1alpha1.VirtualMachine, resSrcVm *res.VirtualMachine) ([]vimTypes.BaseVirtualDeviceConfigSpec, error) {
	netDevices, err := resSrcVm.GetNetworkDevices(ctx)
	if err != nil {
		return nil, err
	}

	var deviceSpecs []vimTypes.BaseVirtualDeviceConfigSpec

	// Note: If no network interface is specified in the vm spec we don't remove existing interfaces while cloning. However,
	// if a default network is configured in vmoperator config then we update the backing for the existing network interfaces.
	if len(vm.Spec.NetworkInterfaces) == 0 {
		if s.network != nil {
			for _, dev := range netDevices {
				backingInfo, err := s.network.EthernetCardBackingInfo(ctx)
				if err != nil {
					return nil, errors.Wrapf(err, "unable to create new ethernet card backing info for network %+v", s.network.Reference())
				}
				dev.GetVirtualDevice().Backing = backingInfo
				deviceSpecs = append(deviceSpecs, &vimTypes.VirtualDeviceConfigSpec{
					Device:    dev,
					Operation: vimTypes.VirtualDeviceConfigSpecOperationEdit,
				})
			}
		}
		return deviceSpecs, nil
	}

	// Remove any existing NICs
	for _, dev := range netDevices {
		deviceSpecs = append(deviceSpecs, &vimTypes.VirtualDeviceConfigSpec{
			Device:    dev,
			Operation: vimTypes.VirtualDeviceConfigSpecOperationRemove,
		})
	}

	// Add new NICs
	addDeviceSpecs, err := s.deviceSpecsFromVM(ctx, vm)
	if err != nil {
		return nil, err
	}
	deviceSpecs = append(deviceSpecs, addDeviceSpecs...)

	return deviceSpecs, nil
}

func processStorageClass(ctx context.Context, resSrcVM *res.VirtualMachine, profileID string) ([]types.BaseVirtualDeviceConfigSpec, []vimTypes.BaseVirtualMachineProfileSpec, error) {
	if len(profileID) == 0 {
		return nil, nil, nil
	}

	disks, err := resSrcVM.GetVirtualDisks(ctx)
	if err != nil {
		return nil, nil, err
	}
	vdcs, err := disks.ConfigSpec(vimTypes.VirtualDeviceConfigSpecOperationEdit)
	if err != nil {
		return nil, nil, err
	}

	var vmProfile []vimTypes.BaseVirtualMachineProfileSpec
	profileSpec := &vimTypes.VirtualMachineDefinedProfileSpec{ProfileId: profileID}
	vmProfile = append(vmProfile, profileSpec)

	for _, cs := range vdcs {
		cs.GetVirtualDeviceConfigSpec().Profile = vmProfile
		cs.GetVirtualDeviceConfigSpec().FileOperation = ""
	}

	return vdcs, vmProfile, nil
}

func (s *Session) getCloneSpec(ctx context.Context, name string, resSrcVM *res.VirtualMachine,
	vm *v1alpha1.VirtualMachine, vmClassSpec *v1alpha1.VirtualMachineClassSpec,
	vmMetadata vmprovider.VirtualMachineMetadata, profileID string) (*vimTypes.VirtualMachineCloneSpec, error) {

	vdcs, vmProfile, err := processStorageClass(ctx, resSrcVM, profileID)
	if err != nil {
		return nil, err
	}
	deviceSpecs, err := s.deviceChangeSpecs(ctx, vm, resSrcVM)
	if err != nil {
		return nil, err
	}
	deviceSpecs = append(deviceSpecs, vdcs...)

	configSpec, err := s.configSpecFromClassSpec(name, &vm.Spec, vmClassSpec, vmMetadata, nil)

	if err != nil {
		return nil, err
	}

	powerOn := vm.Spec.PowerState == v1alpha1.VirtualMachinePoweredOn
	memory := false // No full memory clones

	cloneSpec := &vimTypes.VirtualMachineCloneSpec{
		Config:  configSpec,
		PowerOn: powerOn,
		Memory:  &memory,
	}

	cloneSpec.Location.Pool = vimTypes.NewReference(s.resourcepool.Reference())
	cloneSpec.Location.Profile = vmProfile
	cloneSpec.Location.DeviceChange = deviceSpecs
	cloneSpec.Location.Folder = vimTypes.NewReference(s.folder.Reference())
	vmRef := &vimTypes.ManagedObjectReference{Type: "VirtualMachine", Value: resSrcVM.ReferenceValue()}
	rSpec, err := computeVMPlacement(ctx, s.cluster, vmRef, cloneSpec, vimTypes.PlacementSpecPlacementTypeClone)
	if err != nil {
		return nil, err
	}
	cloneSpec.Location.Host = rSpec.Host
	cloneSpec.Location.Datastore = rSpec.Datastore
	//cloneSpec.Location.DiskMoveType = string(vimTypes.VirtualMachineRelocateDiskMoveOptionsMoveAllDiskBackingsAndConsolidate)
	return cloneSpec, nil
}

func (s *Session) createVm(ctx context.Context, name string, configSpec *vimTypes.VirtualMachineConfigSpec) (*res.VirtualMachine, error) {
	configSpec.Files = &vimTypes.VirtualMachineFileInfo{
		VmPathName: fmt.Sprintf("[%s]", s.datastore.Name()),
	}
	log.Info("Going to create VM.", "Name", name, "ConfigSpec", *configSpec, "Folder", s.folder.Reference().Value, "ResourcePool", s.resourcepool.Reference().Value)
	resVm := res.NewVMForCreate(name)
	err := resVm.Create(ctx, s.folder, s.resourcepool, configSpec)
	if err != nil {
		return nil, err
	}

	// Power on the VM
	err = resVm.SetPowerState(ctx, v1alpha1.VirtualMachinePoweredOn)
	if err != nil {
		return nil, err
	}

	return resVm, nil
}

func (s *Session) cloneVm(ctx context.Context, resSrcVm *res.VirtualMachine, cloneSpec *vimTypes.VirtualMachineCloneSpec) (*res.VirtualMachine, error) {
	log.Info("Going to clone VM", "Name", cloneSpec.Config.Name, "Location", cloneSpec.Location)

	cloneResVm, err := resSrcVm.Clone(ctx, s.folder, cloneSpec)
	if err != nil {
		return nil, err
	}

	return cloneResVm, nil
}

func (s *Session) deployOvf(ctx context.Context, itemID string, vmName string, profileID string) (*res.VirtualMachine, error) {
	var deployment *types.ManagedObjectReference
	var err error
	err = s.WithRestClient(ctx, func(c *rest.Client) error {
		manager := vcenter.NewManager(c)
		dSpec := vcenter.DeploymentSpec{
			Name: vmName,
			// TODO (): Plumb AcceptAllEULA to this Spec
			AcceptAllEULA: true,
		}
		dSpec.StorageProfileID = profileID
		//TODO: Remove this code when storage profile (storageClass) becomes mandatory
		if profileID == "" {
			log.Info("WARNING: ProfileID is empty - using datastore", "datastore", s.datastore.Reference().Value)
			dSpec.DefaultDatastoreID = s.datastore.Reference().Value
		}

		target := vcenter.Target{
			ResourcePoolID: s.resourcepool.Reference().Value,
			FolderID:       s.folder.Reference().Value,
		}

		deploy := vcenter.Deploy{
			DeploymentSpec: dSpec,
			Target:         target,
		}

		deployment, err = manager.DeployLibraryItem(ctx, itemID, deploy)
		return err
	})

	if err != nil {
		return nil, err
	}

	ref, err := s.Finder.ObjectReference(ctx, vimTypes.ManagedObjectReference{Type: deployment.Type, Value: deployment.Value})
	if err != nil {
		return nil, err
	}

	deployedVM, err := res.NewVMFromObject(ref.(*object.VirtualMachine))

	return deployedVM, nil
}

func (s *Session) WithRestClient(ctx context.Context, f func(c *rest.Client) error) error {
	c := rest.NewClient(s.client.VimClient())

	userInfo := url.UserPassword(s.creds.Username, s.creds.Password)

	err := c.Login(ctx, userInfo)
	if err != nil {
		return err
	}

	defer func() {
		if err := c.Logout(ctx); err != nil {
			log.Error(err, "failed to logout")
		}
	}()

	return f(c)
}

func GetExtraConfig(vmSpecMeta, globalMeta map[string]string) []vimTypes.BaseOptionValue {
	var extraConfigs []vimTypes.BaseOptionValue
	mergedConfig := vmSpecMeta

	// If global values for extraConfig have been configured, apply them here
	if globalMeta != nil {
		mergedConfig = make(map[string]string)
		for k, v := range globalMeta {
			mergedConfig[k] = v
		}
		// Ensure that VM-specified extraConfig overrides global values
		for k, v := range vmSpecMeta {
			mergedConfig[k] = v
		}
	}

	for k, v := range mergedConfig {
		extraConfigs = append(extraConfigs, &vimTypes.OptionValue{Key: k, Value: v})
	}
	return extraConfigs
}

func (s *Session) configSpecFromClassSpec(name string, vmSpec *v1alpha1.VirtualMachineSpec, vmClassSpec *v1alpha1.VirtualMachineClassSpec,
	metadata vmprovider.VirtualMachineMetadata, deviceSpecs []vimTypes.BaseVirtualDeviceConfigSpec) (*vimTypes.VirtualMachineConfigSpec, error) {

	configSpec := &vimTypes.VirtualMachineConfigSpec{
		Name:     name,
		NumCPUs:  int32(vmClassSpec.Hardware.Cpus),
		MemoryMB: memoryQuantityToMb(vmClassSpec.Hardware.Memory),
	}

	configSpec.CpuAllocation = &vimTypes.ResourceAllocationInfo{}

	if !vmClassSpec.Policies.Resources.Requests.Cpu.IsZero() {
		rsv := cpuQuantityToMhz(vmClassSpec.Policies.Resources.Requests.Cpu)
		configSpec.CpuAllocation.Reservation = &rsv
	}

	if !vmClassSpec.Policies.Resources.Limits.Cpu.IsZero() {
		lim := cpuQuantityToMhz(vmClassSpec.Policies.Resources.Limits.Cpu)
		configSpec.CpuAllocation.Limit = &lim
	}

	configSpec.MemoryAllocation = &vimTypes.ResourceAllocationInfo{}

	if !vmClassSpec.Policies.Resources.Requests.Memory.IsZero() {
		rsv := memoryQuantityToMb(vmClassSpec.Policies.Resources.Requests.Memory)
		configSpec.MemoryAllocation.Reservation = &rsv
	}

	if !vmClassSpec.Policies.Resources.Limits.Memory.IsZero() {
		lim := memoryQuantityToMb(vmClassSpec.Policies.Resources.Limits.Memory)
		configSpec.MemoryAllocation.Limit = &lim
	}

	if vmSpec.VmMetadata != nil {
		switch vmSpec.VmMetadata.Transport {
		case "ExtraConfig":
			configSpec.ExtraConfig = GetExtraConfig(metadata, s.extraConfig)
		default:
			return nil, fmt.Errorf("unsupported metadata transport %q", vmSpec.VmMetadata.Transport)
		}
	}

	configSpec.Annotation = fmt.Sprint("Virtual Machine managed by VM Operator")

	configSpec.DeviceChange = deviceSpecs

	return configSpec, nil
}

// GetPool returns resource pool for a given invt path of a moref
func GetResourcePool(ctx context.Context, finder *find.Finder, rp string) (*object.ResourcePool, error) {
	ref := types.ManagedObjectReference{Type: "ResourcePool", Value: rp}
	if o, err := finder.ObjectReference(ctx, ref); err == nil {
		return o.(*object.ResourcePool), nil
	}
	return finder.ResourcePool(ctx, rp)
}

// GetPool returns VM folder for a given invt path of a moref
func GetVMFolder(ctx context.Context, finder *find.Finder, folder string) (*object.Folder, error) {
	ref := types.ManagedObjectReference{Type: "Folder", Value: folder}
	if o, err := finder.ObjectReference(ctx, ref); err == nil {
		return o.(*object.Folder), nil
	}
	return finder.Folder(ctx, folder)
}

func IsSupportedDeployType(t string) bool {
	switch t {
	case
		//"vmtx",
		"ovf":
		return true
	}
	return false
}

// getCustomizationSpecs creates the customation spec for the vm
// it is used to config IP for VMs connecting to nsx-t logical ports
func (s *Session) getCustomizationSpecs(namespace, vmName string, vmSpec *v1alpha1.VirtualMachineSpec) (*vimTypes.CustomizationSpec, error) {
	vnifs := []*ncpv1alpha1.VirtualNetworkInterface{}
	np := NsxtNetworkProvider(s.Finder, s.ncpClient)
	for _, nif := range vmSpec.NetworkInterfaces {
		if nif.NetworkType == NsxtNetworkType {
			vnetif, err := np.waitForVnetIFStatus(namespace, nif.NetworkName, vmName)
			if err != nil {
				return nil, err
			}
			vnifs = append(vnifs, vnetif)
		}
	}

	if len(vnifs) == 0 {
		return nil, nil
	}

	customSpec := &vimTypes.CustomizationSpec{
		GlobalIPSettings: vimTypes.CustomizationGlobalIPSettings{},
		// This spec is for Linux guest OS
		// Need to change if other guest OS needs to be supported
		Identity: &vimTypes.CustomizationLinuxPrep{
			HostName: &vimTypes.CustomizationFixedName{
				Name: vmName,
			},
			HwClockUTC: vimTypes.NewBool(true),
		},
	}

	for _, vnetif := range vnifs {
		if len(vnetif.Status.IPAddresses) != 1 {
			log.Info("customize vnetif IP address not unique", "vnetif", vnetif)
			continue
		}
		nicMapping := vimTypes.CustomizationAdapterMapping{
			MacAddress: vnetif.Status.MacAddress,
			Adapter: vimTypes.CustomizationIPSettings{
				Ip: &vimTypes.CustomizationFixedIp{
					IpAddress: vnetif.Status.IPAddresses[0].IP,
				},
				SubnetMask: vnetif.Status.IPAddresses[0].SubnetMask,
				Gateway:    []string{vnetif.Status.IPAddresses[0].Gateway},
			},
		}
		customSpec.NicSettingMap = append(customSpec.NicSettingMap, nicMapping)
	}

	return customSpec, nil
}

// getNetworkProviderByType returns the network provider based on network type
func (s *Session) getNetworkProviderByType(networkType string) (NetworkProvider, error) {
	switch networkType {
	case NsxtNetworkType:
		return NsxtNetworkProvider(s.Finder, s.ncpClient), nil
	case "":
		return DefaultNetworkProvider(s.Finder), nil
	}
	return nil, fmt.Errorf("failed to create network provider for network type '%s'", networkType)
}

func (s *Session) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	if s.client != nil {
		sb.WriteString(fmt.Sprintf("client: %v, ", *s.client))
	}
	if !isNilPtr(s.ncpClient) {
		sb.WriteString(fmt.Sprintf("ncpClient: %+v, ", s.ncpClient))
	}
	if s.contentlib != nil {
		sb.WriteString(fmt.Sprintf("contentlib: %+v, ", *s.contentlib))
	}
	sb.WriteString(fmt.Sprintf("datacenter: %s, ", s.datacenter.Reference().Value))
	if s.folder != nil {
		sb.WriteString(fmt.Sprintf("folder: %s, ", s.folder.Reference().Value))
	}
	if s.network != nil {
		sb.WriteString(fmt.Sprintf("network: %s, ", s.network.Reference().Value))
	}
	if s.resourcepool != nil {
		sb.WriteString(fmt.Sprintf("resourcepool: %s, ", s.resourcepool.Reference().Value))
	}
	if s.cluster != nil {
		sb.WriteString(fmt.Sprintf("cluster: %s, ", s.cluster.Reference().Value))
	}
	if s.datastore != nil {
		sb.WriteString(fmt.Sprintf("datastore: %s ", s.datastore.Reference().Value))
	}
	sb.WriteString("}")
	return sb.String()
}

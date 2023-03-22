// Copyright (c) 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/vm-operator/api/v1alpha2/common"
)

// VirtualMachineNetworkRouteSpec defines a static route for a guest.
type VirtualMachineNetworkRouteSpec struct {
	// To is an IP4 address.
	To string `json:"to"`

	// Via is an IP4 address.
	Via string `json:"via"`

	// Metric is the weight/priority of the route.
	Metric int32 `json:"metric"`
}

// VirtualMachineNetworkInterfaceSpec describes the desired state of a VM's
// network interface.
type VirtualMachineNetworkInterfaceSpec struct {
	// Name describes the unique name of this network interface, used to
	// distinguish it from other network interfaces attached to this VM.
	//
	// This value is also used to rename the device inside the guest when the
	// bootstrap provider is CloudInit. Please note it is up to the user to
	// ensure the provided device name does not conflict with any other devices
	// inside the guest, ex. dvd, cdrom, sda, etc.
	//
	// +kubebuilder:validation:Pattern=^\w\w+$
	Name string `json:"name"`

	// Network is the name of the network resource to which this interface is
	// connected.
	//
	// If no network is provided, then this interface will be connected to the
	// Namespace's default network.
	//
	// +optional
	Network common.PartialObjectRef `json:"network,omitempty"`

	// Addresses is an optional list of IP4 or IP6 addresses to assign to this
	// interface.
	//
	// Please note this field is only supported if the connected network
	// supports manual IP allocation.
	//
	// Please note IP4 and IP6 addresses must include the network prefix length,
	// ex. 192.168.0.10/24 or 2001:db8:101::a/64.
	//
	// Please note this field may not contain IP4 addresses if DHCP4 is set
	// to true or IP6 addresses if DHCP6 is set to true.
	//
	// Please note if the Interfaces field is non-empty then this field is
	// ignored and should be specified on the elements in the Interfaces list.
	//
	// +optional
	Addresses []string `json:"addresses,omitempty"`

	// DHCP4 indicates whether or not this interface uses DHCP for IP4
	// networking.
	//
	// Please note this field is only supported if the network connection
	// supports DHCP.
	//
	// Please note this field is mutually exclusive with IP4 addresses in the
	// Addresses field and the Gateway4 field.
	//
	// +optional
	DHCP4 bool `json:"dhcp4,omitempty"`

	// DHCP6 indicates whether or not this interface uses DHCP for IP6
	// networking.
	//
	// Please note this field is only supported if the network connection
	// supports DHCP.
	//
	// Please note this field is mutually exclusive with IP4 addresses in the
	// Addresses field and the Gateway6 field.
	//
	// +optional
	DHCP6 bool `json:"dhcp6,omitempty"`

	// Gateway4 is the default, IP4 gateway for this interface.
	//
	// Please note this field is only supported if the network connection
	// supports manual IP allocation.
	//
	// If the network connection supports manual IP allocation and the
	// Addresses field includes at least one IP4 address, then this field
	// is required.
	//
	// Please note the IP address must include the network prefix length, ex.
	// 192.168.0.1/24.
	//
	// Please note this field is mutually exclusive with DHCP4.
	//
	// +optional
	Gateway4 string `json:"gateway4,omitempty"`

	// Gateway6 is the primary IP6 gateway for this interface.
	//
	// Please note this field is only supported if the network connection
	// supports manual IP allocation.
	//
	// If the network connection supports manual IP allocation and the
	// Addresses field includes at least one IP4 address, then this field
	// is required.
	//
	// Please note the IP address must include the network prefix length, ex.
	// 2001:db8:101::1/64.
	//
	// Please note this field is mutually exclusive with DHCP6.
	//
	// +optional
	Gateway6 string `json:"gateway6,omitempty"`

	// MTU is the Maximum Transmission Unit size in bytes.
	//
	// Please note this feature is available only with the following bootstrap
	// providers: CloudInit.
	//
	// +optional
	MTU *int64 `json:"mtu,omitempty"`

	// Nameservers is a list of IP4 and/or IP6 addresses used as DNS
	// nameservers.
	//
	// Please note this feature is available only with the following bootstrap
	// providers: CloudInit, LinuxPrep, and Sysprep.
	//
	// Please note that Linux allows only three nameservers
	// (https://linux.die.net/man/5/resolv.conf).
	//
	// +optional
	Nameservers []string `json:"nameservers,omitempty"`

	// Routes is a list of optional, static routes.
	//
	// Please note this feature is available only with the following bootstrap
	// providers: CloudInit.
	//
	// +optional
	Routes []VirtualMachineNetworkRouteSpec `json:"routes,omitempty"`

	// SearchDomains is a list of search domains used when resolving IP
	// addresses with DNS.
	//
	// Please note this feature is available only with the following bootstrap
	// providers: CloudInit, LinuxPrep, and Sysprep.
	//
	// +optional
	SearchDomains []string `json:"searchDomains,omitempty"`
}

// VirtualMachineNetworkSpec defines a VM's desired network configuration.
type VirtualMachineNetworkSpec struct {
	// Network is the optional name of the network resource to which this
	// VM is connected.
	//
	// Please note if the Interfaces field is non-empty then this field is
	// ignored.
	//
	// If networking is not disabled, no interfaces are defined, and this value
	// is omitted, then the VM will be provided a single virtual network
	// interface and connected to the Namespace's default network.
	//
	// +optional
	Network *common.PartialObjectRef `json:"network,omitempty"`

	// Disabled is a flag that indicates whether or not to disable networking
	// for this VM.
	//
	// When set to true, the VM is not configured with a default interface nor
	// any specified from the Interfaces field.
	//
	// +optional
	Disabled bool `json:"disabled,omitempty"`

	// HostName is the value the guest uses as its host name.
	// If omitted then the name of the VM will be used.
	//
	// Please note this feature is available only with the following bootstrap
	// providers: CloudInit, LinuxPrep, and Sysprep.
	//
	// +optional
	HostName string `json:"hostName,omitempty"`

	// Interfaces is the list of network interfaces used by this VM.
	//
	// Please note this field is mutually exclusive with the following fields:
	// DeviceName, Network, Addresses, DHCP4, DHCP6, Gateway4,
	// Gateway6, MTU, Nameservers, Routes, and SearchDomains.
	//
	// +optional
	// +listType=map
	// +listMapKey=name
	Interfaces []VirtualMachineNetworkInterfaceSpec `json:"interfaces,omitempty"`

	// DeviceName describes the unique name of this network interface, used to
	// distinguish it from other network interfaces attached to this VM.
	//
	// This value is also used to rename the device inside the guest when the
	// bootstrap provider is CloudInit. Please note it is up to the user to
	// ensure the provided device name does not conflict with any other devices
	// inside the guest, ex. dvd, cdrom, sda, etc.
	//
	// Please note if the Interfaces field is non-empty then this field is
	// ignored and should be specified on the elements in the Interfaces list.
	//
	// If the Interfaces field is empty and this field is not specified, then
	// the default interface's name will be eth0.
	//
	// +optional
	// +kubebuilder:validation:Pattern=^\w\w+$
	DeviceName string `json:"deviceName,omitempty"`

	// Addresses is an optional list of IP4 or IP6 addresses to assign to the
	// VM.
	//
	// Please note this field is only supported if the connected network
	// supports manual IP allocation.
	//
	// Please note IP4 and IP6 addresses must include the network prefix length,
	// ex. 192.168.0.10/24 or 2001:db8:101::a/64.
	//
	// Please note this field may not contain IP4 addresses if DHCP4 is set
	// to true or IP6 addresses if DHCP6 is set to true.
	//
	// Please note if the Interfaces field is non-empty then this field is
	// ignored and should be specified on the elements in the Interfaces list.
	//
	// +optional
	Addresses []string `json:"addresses,omitempty"`

	// DHCP4 indicates whether or not to use DHCP for IP4 networking.
	//
	// Please note this field is only supported if the network connection
	// supports DHCP.
	//
	// Please note this field is mutually exclusive with IP4 addresses in the
	// Addresses field and the Gateway4 field.
	//
	// Please note if the Interfaces field is non-empty then this field is
	// ignored and should be specified on the elements in the Interfaces list.
	//
	// +optional
	DHCP4 bool `json:"dhcp4,omitempty"`

	// DHCP6 indicates whether or not to use DHCP for IP6 networking.
	//
	// Please note this field is only supported if the network connection
	// supports DHCP.
	//
	// Please note this field is mutually exclusive with IP4 addresses in the
	// Addresses field and the Gateway6 field.
	//
	// Please note if the Interfaces field is non-empty then this field is
	// ignored and should be specified on the elements in the Interfaces list.
	//
	// +optional
	DHCP6 bool `json:"dhcp6,omitempty"`

	// Gateway4 is the default, IP4 gateway for this VM.
	//
	// Please note this field is only supported if the network connection
	// supports manual IP allocation.
	//
	// If the network connection supports manual IP allocation and the
	// Addresses field includes at least one IP4 address, then this field
	// is required.
	//
	// Please note the IP address must include the network prefix length, ex.
	// 192.168.0.1/24.
	//
	// Please note this field is mutually exclusive with DHCP4.
	//
	// Please note if the Interfaces field is non-empty then this field is
	// ignored and should be specified on the elements in the Interfaces list.
	//
	// +optional
	Gateway4 string `json:"gateway4,omitempty"`

	// Gateway6 is the primary IP6 gateway for this VM.
	//
	// Please note this field is only supported if the network connection
	// supports manual IP allocation.
	//
	// If the network connection supports manual IP allocation and the
	// Addresses field includes at least one IP4 address, then this field
	// is required.
	//
	// Please note the IP address must include the network prefix length, ex.
	// 2001:db8:101::1/64.
	//
	// Please note this field is mutually exclusive with DHCP6.
	//
	// Please note if the Interfaces field is non-empty then this field is
	// ignored and should be specified on the elements in the Interfaces list.
	//
	// +optional
	Gateway6 string `json:"gateway6,omitempty"`

	// MTU is the Maximum Transmission Unit size in bytes.
	//
	// Please note this feature is available only with the following bootstrap
	// providers: CloudInit.
	//
	// Please note if the Interfaces field is non-empty then this field is
	// ignored and should be specified on the elements in the Interfaces list.
	//
	// +optional
	MTU *int64 `json:"mtu,omitempty"`

	// Nameservers is a list of IP4 and/or IP6 addresses used as DNS
	// nameservers.
	//
	// Please note this feature is available only with the following bootstrap
	// providers: CloudInit, LinuxPrep, and Sysprep.
	//
	// Please note that Linux allows only three nameservers
	// (https://linux.die.net/man/5/resolv.conf).
	//
	// Please note if the Interfaces field is non-empty then this field is
	// ignored and should be specified on the elements in the Interfaces list.
	//
	// +optional
	Nameservers []string `json:"nameservers,omitempty"`

	// Routes is a list of optional, static routes.
	//
	// Please note this feature is available only with the following bootstrap
	// providers: CloudInit.
	//
	// Please note if the Interfaces field is non-empty then this field is
	// ignored and should be specified on the elements in the Interfaces list.
	//
	// +optional
	Routes []VirtualMachineNetworkRouteSpec `json:"routes,omitempty"`

	// SearchDomains is a list of search domains used when resolving IP
	// addresses with DNS.
	//
	// Please note this feature is available only with the following bootstrap
	// providers: CloudInit, LinuxPrep, and Sysprep.
	//
	// Please note if the Interfaces field is non-empty then this field is
	// ignored and should be specified on the elements in the Interfaces list.
	//
	// +optional
	SearchDomains []string `json:"searchDomains,omitempty"`
}

// VirtualMachineNetworkDNSStatus describes the observed state of the guest's
// RFC 1034 client-side DNS settings.
type VirtualMachineNetworkDNSStatus struct {
	// DHCP indicates whether or not dynamic host control protocol (DHCP) was
	// used to configure DNS configuration.
	//
	// +optional
	DHCP bool `json:"dhcp,omitempty"`

	// DomainName is the domain name portion of the DNS name. For example,
	// the "domain.local" part of "my-vm.domain.local".
	//
	// +optional
	DomainName string `json:"domainName,omitempty"`

	// HostName is the host name portion of the DNS name. For example,
	// the "my-vm" part of "my-vm.domain.local".
	//
	// +optional
	HostName string `json:"hostName,omitempty"`

	// Nameservers is a list of the IP addresses for the DNS servers to use.
	//
	// IP4 addresses are specified using dotted decimal notation. For example,
	// "192.0.2.1".
	//
	// IP6 addresses are 128-bit addresses represented as eight fields of up to
	// four hexadecimal digits. A colon separates each field (:). For example,
	// 2001:DB8:101::230:6eff:fe04:d9ff. The address can also consist of the
	// symbol '::' to represent multiple 16-bit groups of contiguous 0's only
	// once in an address as described in RFC 2373.
	//
	// +optional
	Nameservers []string `json:"nameservers,omitempty"`

	// SearchDomains is a list of domains in which to search for hosts, in the
	// order of preference.
	//
	// +optional
	SearchDomains []string `json:"searchDomains,omitempty"`
}

// VirtualMachineNetworkDHCPOptionsStatus describes the observed state of
// DHCP options.
type VirtualMachineNetworkDHCPOptionsStatus struct {
	// Config describes platform-dependent settings for the DHCP client.
	//
	// The key part is a unique number while the value part is the platform
	// specific configuration command. For example on Linux and BSD systems
	// using the file dhclient.conf output would be reported at system scope:
	// key='1', value='timeout 60;' key='2', value='reboot 10;'. The output
	// reported per interface would be:
	// key='1', value='prepend domain-name-servers 192.0.2.1;'
	// key='2', value='require subnet-mask, domain-name-servers;'.
	//
	// +optional
	// +listType=map
	// +listMapKey=key
	Config []common.KeyValuePair `json:"config,omitempty"`

	// Enabled reports the status of the DHCP client services.
	// +omitempty
	Enabled bool `json:"enabled,omitempty"`
}

// VirtualMachineNetworkDHCPStatus describes the observed state of the
// client-side, system-wide DHCP settings for IP4 and IP6.
type VirtualMachineNetworkDHCPStatus struct {

	// IP4 describes the observed state of the IP4 DHCP client settings.
	//
	// +optional
	IP4 VirtualMachineNetworkDHCPOptionsStatus `json:"ip4,omitempty"`

	// IP6 describes the observed state of the IP6 DHCP client settings.
	//
	// +optional
	IP6 VirtualMachineNetworkDHCPOptionsStatus `json:"ip6,omitempty"`
}

// VirtualMachineNetworkIPRouteGatewayStatus describes the observed state of
// a guest network's IP route's next hop gateway.
type VirtualMachineNetworkIPRouteGatewayStatus struct {
	// Device is the name of the device in the guest for which this gateway
	// applies.
	//
	// +optional
	Device string `json:"device,omitempty"`

	// Address is the IP4 or IP6 address of the gateway.
	//
	// +optional
	Address string `json:"address,omitempty"`
}

// VirtualMachineNetworkIPRouteStatus describes the observed state of a
// guest network's IP routes.
type VirtualMachineNetworkIPRouteStatus struct {
	// Gateway describes where to send the packets to next.
	Gateway VirtualMachineNetworkIPRouteGatewayStatus `json:"gateway"`

	// NetworkAddress is the IP4 or IP6 address of the destination network.
	//
	// Addresses include the network's prefix length, ex. 192.168.0.0/24 or
	// 2001:DB8:101::230:6eff:fe04:d9ff::/64.
	//
	// IP6 addresses are 128-bit addresses represented as eight fields of up to
	// four hexadecimal digits. A colon separates each field (:). For example,
	// 2001:DB8:101::230:6eff:fe04:d9ff. The address can also consist of symbol
	// '::' to represent multiple 16-bit groups of contiguous 0's only once in
	// an address as described in RFC 2373.
	NetworkAddress string `json:"networkAddress"`
}

// VirtualMachineNetworkRouteStatus describes the observed state of a
// guest network's routes.
type VirtualMachineNetworkRouteStatus struct {
	// IPRoutes contain the VM's routing tables for all address families.
	//
	// +optional
	IPRoutes []VirtualMachineNetworkIPRouteStatus `json:"ipRoutes,omitempty"`
}

// VirtualMachineNetworkInterfaceIPAddrStatus describes information about a
// specific IP address.
type VirtualMachineNetworkInterfaceIPAddrStatus struct {
	// Address is an IP4 or IP6 address and their network prefix length.
	//
	// An IP4 address is specified using dotted decimal notation. For example,
	// "192.0.2.1".
	//
	// IP6 addresses are 128-bit addresses represented as eight fields of up to
	// four hexadecimal digits. A colon separates each field (:). For example,
	// 2001:DB8:101::230:6eff:fe04:d9ff. The address can also consist of the
	// symbol '::' to represent multiple 16-bit groups of contiguous 0's only
	// once in an address as described in RFC 2373.
	Address string `json:"address"`

	// Lifetime describes when this address will expire.
	//
	// +optional
	Lifetime metav1.Time `json:"lifetime,omitempty"`

	// Origin describes how this address was configured.
	//
	// +optional
	// +kubebuilder:validation:Enum=dhcp;linklayer;manual;other;random
	Origin string `json:"origin,omitempty"`

	// State describes the state of this IP address.
	//
	// +optional
	// +kubebuilder:validation:Enum=deprecated;duplicate;inaccessible;invalid;preferred;tentative;unknown
	State string `json:"state,omitempty"`
}

// VirtualMachineNetworkInterfaceIPStatus describes the observed state of a
// VM's network interface's IP configuration.
type VirtualMachineNetworkInterfaceIPStatus struct {
	// AutoConfigurationEnabled describes whether or not ICMPv6 router
	// solicitation requests are enabled or disabled from a given interface.
	//
	// These requests acquire an IP6 address and default gateway route from
	// zero-to-many routers on the connected network.
	//
	// If not set then ICMPv6 is not available on this VM.
	//
	// +optional
	AutoConfigurationEnabled *bool `json:"autoConfigurationEnabled,omitempty"`

	// DHCP describes the VM's observed, client-side, interface-specific DHCP
	// options.
	//
	// +optional
	DHCP VirtualMachineNetworkDHCPStatus `json:"dhcp,omitempty"`

	// Addresses describes observed IP addresses for this interface.
	//
	// +optional
	Addresses []VirtualMachineNetworkInterfaceIPAddrStatus `json:"addresses,omitempty"`

	// MACAddr describes the observed MAC address for this interface.
	//
	// +optional
	MACAddr string `json:"macAddr,omitempty"`
}

// VirtualMachineNetworkInterfaceStatus describes the observed state of a
// VM's network interface.
type VirtualMachineNetworkInterfaceStatus struct {
	// Name describes the unique name of this network interface, used to
	// distinguish it from other network interfaces attached to this VM.
	//
	// Please note this name is not related to the name of the device as it is
	// surfaced inside of the guest.
	Name string `json:"name"`

	// IP describes the observed state of the interface's IP configuration.
	//
	// +optional
	IP VirtualMachineNetworkInterfaceIPStatus `json:"ip,omitempty"`

	// DNS describes the observed state of the interface's DNS configuration.
	//
	// +optional
	DNS VirtualMachineNetworkDNSStatus `json:"dns,omitempty"`
}

// VirtualMachineNetworkIPStackStatus describes the observed state of a
// a VM's IP stack.
type VirtualMachineNetworkIPStackStatus struct {
	// DHCP describes the VM's observed, client-side, system-wide DHCP options.
	//
	// +optional
	DHCP VirtualMachineNetworkDHCPStatus `json:"dhcp,omitempty"`

	// DNS describes the VM's observed, client-side DNS configuration.
	//
	// +optional
	DNS VirtualMachineNetworkDNSStatus `json:"dns,omitempty"`

	// IPRoutes contain the VM's routing tables for all address families.
	//
	// +optional
	IPRoutes []VirtualMachineNetworkIPRouteStatus `json:"ipRoutes,omitempty"`

	// KernelConfig describes the observed state of the VM's kernel IP
	// configuration settings.
	//
	// The key part contains a unique number while the value part contains the
	// 'key=value' as provided by the underlying provider. For example, on
	// Linux and/or BSD, the systcl -a output would be reported as:
	// key='5', value='net.ipv4.tcp_keepalive_time = 7200'.
	//
	// +optional
	// +listType=map
	// +listMapKey=key
	KernelConfig []common.KeyValuePair `json:"kernelConfig,omitempty"`
}

// VirtualMachineNetworkStatus defines the observed state of a VM's
// network configuration.
type VirtualMachineNetworkStatus struct {

	// Interfaces describes the status of the VM's network interfaces.
	//
	// +optional
	// +listType=map
	// +listMapKey=name
	Interfaces []VirtualMachineNetworkInterfaceStatus `json:"interfaces,omitempty"`

	// PrimaryIP4 describes the VM's primary IP4 address.
	//
	// If the bootstrap provider is CloudInit then this value is set to the
	// value of the VM's "guestinfo.local-ipv4" property. Please see
	// https://bit.ly/3A66vZg for more information on how this value is
	// calculated.
	//
	// If the bootstrap provider is anything else then this field is set to the
	// the value of the infrastructure VM's "guest.ipAddress" field. Please see
	// https://bit.ly/3Au0jM4 for more information.
	//
	// +optional
	PrimaryIP4 string `json:"primaryIP4,omitempty"`

	// PrimaryIP6 describes the VM's primary IP6 address.
	//
	// If the bootstrap provider is CloudInit then this value is set to the
	// value of the VM's "guestinfo.local-ipv6" property. Please see
	// https://bit.ly/3A66vZg for more information on how this value is
	// calculated.
	//
	// If the bootstrap provider is anything else then this field is set to the
	// the value of the infrastructure VM's "guest.ipAddress" field. Please see
	// https://bit.ly/3Au0jM4 for more information.
	//
	// +optional
	PrimaryIP6 string `json:"primaryIP6,omitempty"`

	VirtualMachineNetworkIPStackStatus `json:",inline"`
}

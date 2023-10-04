// Copyright (c) 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// +kubebuilder:object:generate=true

package sysprep

import (
	corev1 "k8s.io/api/core/v1"
)

// Sysprep describes the object representation of a Windows sysprep.xml answer
// file.
//
// All fields and their values are transferred into the sysprep.xml file that
// VirtualCenter stores on the target virtual disk.
//
// For more detailed information, please see
// https://technet.microsoft.com/en-us/library/cc771830(v=ws.10).aspx.
type Sysprep struct {

	// GUIRunOnce is a representation of the Sysprep GuiRunOnce key.
	//
	// +optional
	GUIRunOnce GUIRunOnce `json:"guiRunOnce,omitempty"`

	// GUIUnattended is a representation of the Sysprep GUIUnattended key.
	GUIUnattended GUIUnattended `json:"guiUnattended"`

	// Identification is a representation of the Sysprep Identification key.
	Identification Identification `json:"identification"`

	// LicenseFilePrintData is a representation of the Sysprep
	// LicenseFilePrintData key.
	//
	// Please note this is required only for Windows 2000 Server and Windows
	// Server 2003.
	//
	// +optional
	LicenseFilePrintData *LicenseFilePrintData `json:"licenseFilePrintData,omitempty"`

	// UserData is a representation of the Sysprep UserData key.
	UserData UserData `json:"userData"`
}

// GUIRunOnce maps to the GuiRunOnce key in the sysprep.xml answer file.
type GUIRunOnce struct {
	// Commands is a list of commands to run at first user logon, after guest
	// customization.
	//
	// +optional
	Commands []string `json:"commands,omitempty"`
}

// GUIUnattended maps to the GuiUnattended key in the sysprep.xml answer file.
type GUIUnattended struct {

	// AutoLogon determine whether or not the machine automatically logs on as
	// Administrator.
	//
	// Please note if AutoLogin is true, then Password must be set or guest
	// customization will fail.
	//
	// +optional
	AutoLogon bool `json:"autoLogon,omitempty"`

	// AutoLogonCount specifies the number of times the machine should
	// automatically log on as Administrator.
	//
	// Generally it should be 1, but if your setup requires a number of reboots,
	// you may want to increase it. This number may be determined by the list of
	// commands executed by the GuiRunOnce command.
	//
	// Please note this field only matters if AutoLogin is true.
	//
	// +optional
	AutoLogonCount int32 `json:"autoLogonCount,omitempty"`

	// Password is the new administrator password for the machine.
	//
	// To specify that the password should be set to blank (that is, no
	// password), set the password value to NULL. Because of encryption, "" is
	// NOT a valid value.
	//
	// Please note if the password is set to blank and AutoLogon is true, the
	// guest customization will fail.
	//
	// If the XML file is generated by the VirtualCenter Customization Wizard,
	// then the password is encrypted. Otherwise, the client should set the
	// plainText attribute to true, so that the customization process does not
	// attempt to decrypt the string.
	//
	// +optional
	Password corev1.SecretKeySelector `json:"password,omitempty"`

	// TimeZone is the time zone index for the virtual machine.
	//
	// Please note that numbers correspond to time zones listed at
	// https://bit.ly/3Rzv8oL.
	//
	// +optional
	TimeZone int32 `json:"timeZone,omitempty"`
}

// Identification maps to the Identification key in the sysprep.xml answer file
// and provides information needed to join a workgroup or domain.
type Identification struct {

	// DomainAdmin is the domain user account used for authentication if the
	// virtual machine is joining a domain. The user does not need to be a
	// domain administrator, but the account must have the privileges required
	// to add computers to the domain.
	//
	// +optional
	DomainAdmin string `json:"domainAdmin,omitempty"`

	// DomainAdminPassword is the password for the domain user account used for
	// authentication if the virtual machine is joining a domain.
	//
	// +optional
	DomainAdminPassword corev1.SecretKeySelector `json:"domainAdminPassword,omitempty"`

	// JoinDomain is the domain that the virtual machine should join. If this
	// value is supplied, then DomainAdmin and DomainAdminPassword must also be
	// supplied, and the JoinWorkgroup name must be empty.
	//
	// +optional
	JoinDomain string `json:"joinDomain,omitempty"`

	// JoinWorkgroup is the workgroup that the virtual machine should join. If
	// this value is supplied, then the JoinDomain and the authentication fields
	// (DomainAdmin and DomainAdminPassword) must be empty.
	//
	// +optional
	JoinWorkgroup string `json:"joinWorkgroup,omitempty"`
}

// CustomizationLicenseDataMode is an enumeration of the different license
// modes.
//
// +kubebuilder:validation:Enum=perSeat;perServer
type CustomizationLicenseDataMode string

const (
	// CustomizationLicenseDataModePerSeat indicates that a client access
	// license has been purchased for each computer that accesses the
	// VirtualCenter server.
	CustomizationLicenseDataModePerSeat CustomizationLicenseDataMode = "perSeat"

	// CustomizationLicenseDataModePerServer indicates that client access
	// licenses have been purchased for the server, allowing a certain number of
	// concurrent connections to the VirtualCenter server.
	CustomizationLicenseDataModePerServer CustomizationLicenseDataMode = "perServer"
)

// LicenseFilePrintData maps to the LicenseFilePrintData key in the sysprep.xml
// answer file and provides information needed to join a workgroup or domain.
type LicenseFilePrintData struct {

	// AutoMode specifies the server licensing mode.
	AutoMode CustomizationLicenseDataMode `json:"autoMode"`

	// AutoUsers indicates the number of client licenses purchased for the
	// VirtualCenter server being installed.
	//
	// Please note this value is ignored unless AutoMode is PerServer.
	//
	// +optional
	AutoUsers *int32 `json:"autoUsers,omitempty"`
}

// UserData maps to the UserData key in the sysprep.xml answer file and provides
// personal data pertaining to the owner of the virtual machine.
type UserData struct {

	// FullName is the user's full name.
	//
	// +optional
	FullName string `json:"fullName,omitempty"`

	// OrgName is the name of the user's organization.
	//
	// +optional
	OrgName string `json:"orgName,omitempty"`

	// ProductID is a valid serial number.
	//
	// Please note unless the VirtualMachineImage was installed with a volume
	// license key, ProductID must be set or guest customization will fail.
	//
	// +optional
	ProductID corev1.SecretKeySelector `json:"productID,omitempty"`
}

// Copyright (c) 2024 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	"unsafe"

	apiconversion "k8s.io/apimachinery/pkg/conversion"

	vmopv1a2sysprep "github.com/vmware-tanzu/vm-operator/api/v1alpha2/sysprep"
	vmopv1sysprep "github.com/vmware-tanzu/vm-operator/api/v1alpha3/sysprep"
)

// Please see https://github.com/kubernetes/code-generator/issues/172 for why
// this function exists in this directory structure.
func Convert_sysprep_Sysprep_To_sysprep_Sysprep(
	in *vmopv1sysprep.Sysprep, out *vmopv1a2sysprep.Sysprep, s apiconversion.Scope) error {

	if in.GUIRunOnce != nil {
		out.GUIRunOnce = *(*vmopv1a2sysprep.GUIRunOnce)(unsafe.Pointer(in.GUIRunOnce))
	}
	out.GUIUnattended = (*vmopv1a2sysprep.GUIUnattended)(unsafe.Pointer(in.GUIUnattended))
	out.LicenseFilePrintData = (*vmopv1a2sysprep.LicenseFilePrintData)(unsafe.Pointer(in.LicenseFilePrintData))
	out.UserData = (*vmopv1a2sysprep.UserData)(unsafe.Pointer(in.UserData))
	if id := in.Identification; id != nil {
		out.Identification = &vmopv1a2sysprep.Identification{
			DomainAdmin:         id.DomainAdmin,
			DomainAdminPassword: (*vmopv1a2sysprep.DomainPasswordSecretKeySelector)(unsafe.Pointer(id.DomainAdminPassword)),
			JoinWorkgroup:       id.JoinWorkgroup,
		}
	}

	return nil
}

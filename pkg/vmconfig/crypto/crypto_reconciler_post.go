// Copyright (c) 2024 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"context"

	"github.com/vmware/govmomi/fault"
	"github.com/vmware/govmomi/vim25/mo"
	vimtypes "github.com/vmware/govmomi/vim25/types"

	"github.com/vmware-tanzu/vm-operator/pkg/conditions"
	"github.com/vmware-tanzu/vm-operator/pkg/vmconfig/crypto/internal"

	vmopv1 "github.com/vmware-tanzu/vm-operator/api/v1alpha3"
)

//nolint:gocyclo
func (r reconciler) OnResult(
	ctx context.Context,
	vm *vmopv1.VirtualMachine,
	moVM mo.VirtualMachine,
	resultErr error) error {

	if ctx == nil {
		panic("context is nil")
	}
	if moVM.Config == nil {
		panic("moVM.config is nil")
	}
	if vm == nil {
		panic("vm is nil")
	}

	state := internal.FromContext(ctx)

	// Update the VM's crypto status.
	if key := getCurCryptoKey(moVM); key.provider == "" && key.id == "" {
		vm.Status.Crypto = nil
	} else {
		if vm.Status.Crypto == nil {
			vm.Status.Crypto = &vmopv1.VirtualMachineCryptoStatus{}
		}
		vm.Status.Crypto.ProviderID = key.provider
		vm.Status.Crypto.KeyID = key.id

		if state.IsEncStorClass {
			vm.Status.Crypto.Encrypted = []vmopv1.VirtualMachineEncryptionType{
				vmopv1.VirtualMachineEncryptionTypeConfig,
				vmopv1.VirtualMachineEncryptionTypeDisks,
			}
		} else if ok, _ := hasVTPM(moVM, nil); ok {
			vm.Status.Crypto.Encrypted = []vmopv1.VirtualMachineEncryptionType{
				vmopv1.VirtualMachineEncryptionTypeConfig,
			}
		}
	}

	if resultErr == nil {

		// If no reconfigure error occurred then we need to check if there was
		// a crypto update as part of the reconfigure.

		if state.Operation == "encrypting" || state.Operation == "recrypting" {

			// A crypto update was successful, so indicate that the encryption
			// state of this VM is synced.
			conditions.MarkTrue(vm, vmopv1.VirtualMachineEncryptionSynced)
		}

		return nil
	}

	// Determine the message to put on the condition.
	var msgs []string

	//
	// At this point we know that a reconfigure error occurred *and* there was
	// a crypto update in the ConfigSpec. It is time to parse the reconfigErr
	// to determine if it was related to the crypto update.
	//

	fault.In(
		resultErr,
		func(
			fault vimtypes.BaseMethodFault,
			localizedMessage string,
			localizableMessages []vimtypes.LocalizableMessage) bool {

			switch tErr := fault.(type) {
			case *vimtypes.GenericVmConfigFault:
				for i := range localizableMessages {
					switch localizableMessages[i].Key {
					case "msg.vigor.enc.keyNotFound":
						msgs = append(msgs, "specify a valid key")
					case "msg.keysafe.locator":
						msgs = append(msgs, "specify a key that can be located")
					case "msg.vtpm.add.notEncrypted":
						msgs = append(msgs, "add vTPM")
					case "msg.vigor.enc.required.vtpm":
						msgs = append(msgs, "have vTPM")
					}
				}
			case *vimtypes.SystemError:
				switch localizedMessage {
				case "Error creating disk Key locator":
					msgs = append(msgs, "specify a valid key")
				case "Key locator error":
					msgs = append(msgs, "specify a key that can be located")
				case "Key required for encryption.bundle.":
					msgs = append(msgs, "not specify encryption bundle")
				}
			case *vimtypes.NotSupported:
				for i := range localizableMessages {
					//nolint:gocritic
					switch localizableMessages[i].Key {
					case "msg.disk.policyChangeFailure":
						msgs = append(msgs, "not have encryption IO filter")
					}
				}
			case *vimtypes.InvalidArgument:
				for i := range localizableMessages {
					//nolint:gocritic
					switch localizableMessages[i].Key {
					case "config.extraConfig[\"dataFileKey\"]":
						msgs = append(msgs, "not set secret key")
					}
				}
			case *vimtypes.InvalidDeviceOperation:
				for i := range localizableMessages {
					switch localizableMessages[i].Key {
					case "msg.hostd.deviceSpec.enc.encrypted":
						msgs = append(msgs, "not specify encrypted disk")
					case "msg.hostd.deviceSpec.enc.notEncrypted":
						msgs = append(msgs, "not specify decrypted disk")
					default:
						msgs = append(msgs, "not add/remove device sans crypto spec")
					}
				}

			case *vimtypes.InvalidDeviceSpec:
				for i := range localizableMessages {
					switch localizableMessages[i].Key {
					case "msg.hostd.deviceSpec.enc.badPolicy":
						msgs = append(msgs, "have encryption IO filter")
					case "msg.hostd.deviceSpec.enc.notDisk":
						msgs = append(msgs, "not apply only to disk")
					case "msg.hostd.deviceSpec.enc.sharedBacking":
						msgs = append(msgs, "not have disk with shared backing")
					case "msg.hostd.deviceSpec.enc.notFile":
						msgs = append(msgs, "not have raw disk mapping")
					case "msg.hostd.configSpec.enc.mismatch":
						msgs = append(msgs, "not add encrypted disk")
					case "msg.hostd.deviceSpec.add.noencrypt":
						msgs = append(msgs, "not add plain disk")
					}
				}
			case *vimtypes.InvalidPowerState:
				if tErr.ExistingState != vimtypes.VirtualMachinePowerStatePoweredOff {
					msgs = append(msgs, "be powered off")
				}
			case *vimtypes.InvalidVmConfig:
				for i := range localizableMessages {
					switch localizableMessages[i].Key {
					case "msg.hostd.configSpec.enc.snapshots":
						msgs = append(msgs, "not have snapshots")
					case "msg.hostd.deviceSpec.enc.diskChain":
						msgs = append(msgs, "not have only disk snapshots")
					case "msg.hostd.configSpec.enc.notEncrypted":
						msgs = append(msgs, "not be encrypted")
					case "msg.hostd.configSpec.enc.encrypted":
						msgs = append(msgs, "be encrypted")
					case "msg.hostd.configSpec.enc.mismatch":
						msgs = append(msgs, "have vm and disks with different encryption states")
					}
				}
			}
			return false
		},
	)

	if len(msgs) > 0 {
		markEncryptionStateNotSynced(
			vm,
			state.Operation,
			ReasonReconfigureError,
			msgs...)
	}

	return nil
}
# Deploying a VM

Be our guest, be our guest...

---

## Overview

This page reviews the different components, workflows, and decisions related to deploying a VM with VM Operator:

## The VirtualMachine API

```yaml
apiVersion: vmoperator.vmware.com/v1alpha1 # (1)
kind: VirtualMachine # (2)
metadata:
  name:      my-vm # (3)
  namespace: my-namespace # (4)
spec:
  className:    small # (5)
  imageName:    ubuntu-2210 # (6)
  storageClass: iscsi # (7)
  vmMetadata: # (8)
    transport: CloudInit
    secretName: my-vm-bootstrap-data
```

1.  :wave: The field `apiVersion` indicates the resource's schema, ex. `vmoperator.vmware.com`, and version, ex.`v1alpha1`.

2.  :wave: The field `kind` specifies the kind of resource, ex. `VirtualMachine`.

3.  :wave: The field `metadata.name` is used to uniquely identify an instance of an API resource in a given Kubernetes namespace.

4.  :wave: The field `metadata.namespace` denotes in which Kubernetes namespace the API resource is located.

5.  :wave: The field `spec.className` refers to the name of the `VirtualMachineClass` resource that provides the hardware configuration when deploying a VM.

    The `VirtualMachineClass` API is cluster-scoped, and the following command may be used to print all of the VM classes on a cluster:

    ```shell
    kubectl get vmclass
    ```

    However, access to these resources is per-namespace. To determine the names of the VM classes that may be used in a given namespace, use the following command: 

    ```shell
    kubectl get -n <NAMESPACE> vmclassbinding
    ```

6.  :wave: The field `spec.imageName` refers to the name of the `ClusterVirtualMachineImage` or `VirtualMachineImage` resource that provides the disk(s) when deploying a VM.

    * If there is a `ClusterVirtualMachineImage` resource with the specified name, the cluster-scoped resource is used, otherwise...
    * If there is a `VirtualMachineImage` resource in the same namespace as the VM being deployed, the namespace-scoped resource is used.

    The following command may be used to print a list of the images available to the entire cluster:

    ```shell
    kubectl get clustervmimage
    ```

    Whereas this command may be used to print a list of images available to a given namespace:

    ```shell
    kubectl get -n <NAMESPACE> vmimage
    ```

7.  :wave: The field `spec.storageClass` refers to the Kubernetes [storage class](https://kubernetes.io/docs/concepts/storage/storage-classes/) used to configure the storage for the VM.

    The following command may be used to print a list of the storage classes available to the entire cluster:

    ```shell
    kubectl get storageclass
    ```

8.  :wave: The field `spec.vmMetadata`, and the fields inside of it, are used to configure the VM's [bootstrap provider](#bootstrap-provider).

## Bootstrap Provider

There are a number of methods that may be used to bootstrap a virtual machine's (VM) guest operating system:

| Provider                    | Supported    | Network Config                | Linux | Windows | Description |
|-----------------------------|:------------:|-------------------------------|:-----:|:-------:|-------------|
| [Cloud-Init](#cloud-init)   |       ✓      | [Cloud-Init Network v2](https://cloudinit.readthedocs.io/en/latest/reference/network-config-format-v2.html) |   ✓   |     ✓    | The industry standard, multi-distro method for cross-platform, cloud instance initialization with modern, VM images |
| [vAppConfig](#vappconfig)   |       ✓      | Bespoke                       |   ✓   |         | For images with bespoke, bootstrap engines driven by vAppConfig properties |
| [OvfEnv](#ovfenv)           | _deprecated_ | [Guest OS Customization](https://vdc-download.vmware.com/vmwb-repository/dcr-public/c476b64b-c93c-4b21-9d76-be14da0148f9/04ca12ad-59b9-4e1c-8232-fd3d4276e52c/SDK/vsphere-ws/docs/ReferenceGuide/vim.vm.customization.Specification.html) (GOSC) |   ✓   |         | A combination of GOSC and Cloud-Init user-data |
| [ExtraConfig](#extraconfig) | _deprecated_ | GOSC                          |   ✓   |         | For images with bespoke, bootstrap engines that rely on Guest Info data |

!!! note "`ConfigMap` or `Secret`"

    The choice of a `ConfigMap` or `Secret` resource in no way impacts the choice or behavior of the selected bootstrap provider. When VM Operator was first released, the only way to store bootstrap data was via a `ConfigMap` resource. While this still works, it is not recommended as data stored in a `ConfigMap` is not encrypted at rest. Instead, it is recommended users switch to using `Secret` resources for storing bootstrap data. Aside from how the data is stored in etcd, the following two resources are effectively identical and serve the same purpose:

    === "CloudConfig in a ConfigMap"

        ``` yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name:      my-vm-bootstrap-data
          namespace: my-namespace
        data:
          user-data: |
            #cloud-config
            users:
            - default
        ```

    === "CloudConfig in a Secret"

        ``` yaml
        apiVersion: v1
        kind: Secret
        metadata:
          name:      my-vm-bootstrap-data
          namespace: my-namespace
        stringData:
          user-data: |
            #cloud-config
            users:
            - default
        ```
### Supported

#### Cloud-Init

[Cloud-Init](https://cloudinit.readthedocs.io/en/latest/) is widely recognized as the de facto method for bootstrapping modern VM instances on hyperscalers, including Tanzu with VM Operator. For example, the following resources may be used to deploy a VM and bootstrap its guest with Cloud-Init to:

* add a custom user
* execute commands on boot
* write files

=== "VirtualMachine"

    ``` yaml
    apiVersion: vmoperator.vmware.com/v1alpha1
    kind: VirtualMachine
    metadata:
      name:      my-vm
      namespace: my-namespace
    spec:
      className:    small
      imageName:    ubuntu-2210
      storageClass: iscsi
      vmMetadata:
        transport: CloudInit
        secretName: my-vm-bootstrap-data
    ```

=== "CloudConfig"

    ``` yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name:      my-vm-bootstrap-data
      namespace: my-namespace
    stringData:
      user-data: |
        #cloud-config
        users:
        - default
        - name: akutz
          primary_group: akutz
          groups: users
          ssh_authorized_keys:
          - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDSL7uWGj...
        runcmd:
        - "ls /"
        - [ "ls", "-a", "-l", "/" ]
        write_files:
        - path: /etc/my-plaintext
          permissions: '0644'
          owner: root:root
          content: |
            Hello, world.
    ```

The data in the above `Secret` is called the Cloud-Init _Cloud Config_. For more information on the Cloud-Init Cloud Config format, please see its [official documentation](https://cloudinit.readthedocs.io/en/latest/reference/examples.html).

!!! note "Windows and Cloud-Init"

    It is possible to use the Cloud-Init bootstrap provider to deploy a Windows image if it contains [Cloudbase-Init](https://cloudbase.it/cloudbase-init/), the Windows port of Cloud-Init.

#### vAppConfig

The vAppConfig bootstrap method is useful for legacy, VM images that rely on bespoke, boot-time processes that leverage vAppConfig properties for customizing a guest.

### Deprecated

The following bootstrap providers are still available, but they are deprecated and are not recommended.

#### OvfEnv

The `OvfEnv` method is no longer recommended. It relied on a combination of VMware's Guest OS Customization (GOSC) APIs for bootstrapping the guest's network and the Cloud-Init OVF data source for supplying a Cloud-Init Cloud Config. As a result of mixing bootstrap engines (GOSC and Cloud-Init), there was a race condition that meant any image that used `OvfEnv` needed to have a special fix applied. This provider is no longer supported and will be removed in v1alpha2. Any consumers still relying on this provider should switch to Cloud-Init.

#### ExtraConfig

When Tanzu Kubernetes was first released, the Cluster API provider that depended upon VM Operator used the `ExtraConfig` provider for supplying bootstrap information. This method was never intended for wide use, and Tanzu now uses Cloud-Init anyway. To that end, this provider is no longer supported and will be removed in v1alpha2. Any consumers still relying on this provider should switch to Cloud-Init.
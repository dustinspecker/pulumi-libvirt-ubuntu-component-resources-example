package vm

import (
	"fmt"

	"github.com/pulumi/pulumi-libvirt/sdk/go/libvirt"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type VM struct {
	pulumi.ResourceState

	Name pulumi.StringOutput `pulumi:"name"`
	IP   pulumi.StringOutput `pulumi:"ip"`
}

func NewVM(ctx *pulumi.Context, name string, poolName pulumi.StringOutput, baseDiskID pulumi.IDOutput, cloudInitDiskID pulumi.IDOutput, networkID pulumi.IDOutput, opts ...pulumi.ResourceOption) (*VM, error) {
	// new VM resource to create
	var resource VM

	// register the component
	err := ctx.RegisterComponentResource("pulumi-libvirt-ubuntu:pkg/vm:vm", name, &resource, opts...)
	if err != nil {
		return nil, err
	}

	// create a filesystem volume for our VM
	// This filesystem will be based on the `ubuntu` volume above
	// we'll use a size of 10GB
	filesystem, err := libvirt.NewVolume(ctx, fmt.Sprintf("%s-filesystem", name), &libvirt.VolumeArgs{
		BaseVolumeId: baseDiskID,
		Pool:         poolName,
		Size:         pulumi.Int(10000000000),
	}, pulumi.Parent(&resource))
	if err != nil {
		return nil, err
	}

	// create a VM that has a name starting with ubuntu
	domain, err := libvirt.NewDomain(ctx, fmt.Sprintf("%s-domain", name), &libvirt.DomainArgs{
		Autostart: pulumi.Bool(true),
		Cloudinit: cloudInitDiskID,
		Consoles: libvirt.DomainConsoleArray{
			// enables using `virsh console ...`
			libvirt.DomainConsoleArgs{
				Type:       pulumi.String("pty"),
				TargetPort: pulumi.String("0"),
				TargetType: pulumi.String("serial"),
			},
		},
		Disks: libvirt.DomainDiskArray{
			libvirt.DomainDiskArgs{
				VolumeId: filesystem.ID(),
			},
		},
		NetworkInterfaces: libvirt.DomainNetworkInterfaceArray{
			libvirt.DomainNetworkInterfaceArgs{
				NetworkId:    networkID,
				WaitForLease: pulumi.Bool(true),
			},
		},
		// delete existing VM before creating replacement to avoid two VMs trying to use the same volume
	}, pulumi.Parent(&resource), pulumi.ReplaceOnChanges([]string{"*"}), pulumi.DeleteBeforeReplace(true))

	resource.Name = domain.Name
	resource.IP = domain.NetworkInterfaces.Index(pulumi.Int(0)).Addresses().Index(pulumi.Int(0))
	ctx.RegisterResourceOutputs(&resource, pulumi.Map{
		"ip":   domain.NetworkInterfaces.Index(pulumi.Int(0)).Addresses().Index(pulumi.Int(0)),
		"name": domain.Name,
	})

	return &resource, err
}

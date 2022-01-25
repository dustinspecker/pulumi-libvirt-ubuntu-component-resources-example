package vm

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi-libvirt/sdk/go/libvirt"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type VMGroup struct {
	pulumi.ResourceState

	Name pulumi.String         `pulumi:"name"`
	VMs  pulumi.StringMapArray `pulumi:"vms"`
}

func NewVMGroup(ctx *pulumi.Context, groupName string, hostStoragePoolPath string, vmImageSource string, ipCIDR string, numberOfVMs int, opts ...pulumi.ResourceOption) (*VMGroup, error) {
	var resource VMGroup

	err := ctx.RegisterComponentResource("pulumi-libvirt-ubuntu:pkg/vm:vmgroup", groupName, &resource, opts...)
	if err != nil {
		return nil, err
	}

	// `pool` is a storage pool that can be used to create volumes
	// the `dir` type uses a directory to manage files
	// `Path` maps to a directory on the host filesystem, so we'll be able to
	// volume contents in `/pool/cluster_storage/`
	pool, err := libvirt.NewPool(ctx, fmt.Sprintf("%s-cluster", groupName), &libvirt.PoolArgs{
		Type: pulumi.String("dir"),
		Path: pulumi.String(hostStoragePoolPath),
	}, pulumi.Parent(&resource), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, err
	}

	// create a volume with the contents being a Ubuntu 20.04 server image
	imageVolume, err := libvirt.NewVolume(ctx, fmt.Sprintf("%s-image", groupName), &libvirt.VolumeArgs{
		Pool:   pool.Name,
		Source: pulumi.String(vmImageSource),
	}, pulumi.Parent(&resource))
	if err != nil {
		return nil, err
	}

	cloud_init_user_data, err := os.ReadFile("./cloud_init_user_data.yaml")
	if err != nil {
		return nil, err
	}

	cloud_init_network_config, err := os.ReadFile("./cloud_init_network_config.yaml")
	if err != nil {
		return nil, err
	}

	// create a cloud init disk that will setup the ubuntu credentials
	cloud_init, err := libvirt.NewCloudInitDisk(ctx, fmt.Sprintf("%s-cloud-init", groupName), &libvirt.CloudInitDiskArgs{
		MetaData:      pulumi.String(string(cloud_init_user_data)),
		NetworkConfig: pulumi.String(string(cloud_init_network_config)),
		Pool:          pool.Name,
		UserData:      pulumi.String(string(cloud_init_user_data)),
	}, pulumi.Parent(&resource))
	if err != nil {
		return nil, err
	}

	// create NAT network using 192.168.10/24 CIDR
	network, err := libvirt.NewNetwork(ctx, fmt.Sprintf("%s-network", groupName), &libvirt.NetworkArgs{
		Addresses: pulumi.StringArray{pulumi.String(ipCIDR)},
		Autostart: pulumi.Bool(true),
		Mode:      pulumi.String("nat"),
	}, pulumi.Parent(&resource), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, err
	}

	vmOutputs := pulumi.StringMapArray{}

	for i := 0; i < numberOfVMs; i++ {
		vmName := fmt.Sprintf("%s-%d", groupName, i)

		vm, err := NewVM(ctx, vmName, pool.Name, imageVolume.ID(), cloud_init.ID(), network.ID(), pulumi.Parent(&resource))
		if err != nil {
			return nil, err
		}

		vmOutputs = append(vmOutputs, pulumi.StringMap{
			"ip":   vm.IP,
			"name": vm.Name,
		})
	}

	resource.Name = pulumi.String(groupName)
	resource.VMs = vmOutputs

	ctx.RegisterResourceOutputs(&resource, pulumi.Map{
		"name": pulumi.String(groupName),
		"vms":  vmOutputs,
	})

	return &resource, nil
}

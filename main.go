package main

import (
	"pulumi-libvirt-ubuntu/pkg/vm"

	"github.com/pulumi/pulumi-libvirt/sdk/go/libvirt"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		domainsUse1GBMemory := func(args *pulumi.ResourceTransformationArgs) *pulumi.ResourceTransformationResult {
			// only modify resources that are a Domain type
			if args.Type == "libvirt:index/domain:Domain" {
				modifiedDomainArgs := args.Props.(*libvirt.DomainArgs)
				modifiedDomainArgs.Memory = pulumi.Int(1024)

				return &pulumi.ResourceTransformationResult{
					Props: modifiedDomainArgs,
					Opts:  args.Opts,
				}
			}

			return nil
		}

		vmGroup, err := vm.NewVMGroup(ctx, "ubuntu", "/pool/cluster_storage", "https://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img", "192.168.10.0/24", 3, pulumi.Transformations([]pulumi.ResourceTransformation{domainsUse1GBMemory}))
		if err != nil {
			return err
		}

		ctx.Export("VMs", vmGroup.VMs)

		return nil
	})
}

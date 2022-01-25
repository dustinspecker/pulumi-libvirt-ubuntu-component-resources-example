# pulumi-libvirt-ubuntu-component-resources-example

> Based on https://dustinspecker.com/posts/ubuntu-vm-pulumi-libvirt-component-resources/

## Usage

1. Install [Pulumi](https://www.pulumi.com/)
1. Clone this repository
1. Run `pulumi login`
1. Run `pulumi stack init dev`
1. Run `pulumi config set libvirt:uri qemu:///system`
1. Run `pulumi up`

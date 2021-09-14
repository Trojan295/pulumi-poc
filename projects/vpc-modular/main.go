package main

import (
	"github.com/Trojan295/pulumi-poc/pkg/vpc"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	vpcCidrBlock            = "10.0.0.0/16"
	publicSubnetCidrBlocks  = []string{"10.0.0.0/24", "10.0.1.0/24", "10.0.2.0/24"}
	privateSubnetCidrBlocks = []string{"10.0.100.0/24", "10.0.101.0/24", "10.0.102.0/24"}
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		availabilityZones, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{})
		if err != nil {
			return err
		}

		output, err := vpc.NewVpc(ctx, &vpc.VpcInput{
			VpcCidrBlock:            vpcCidrBlock,
			AvailabilityZones:       availabilityZones.Names,
			PrivateSubnetCidrBlocks: privateSubnetCidrBlocks,
			PublicSubnetCidrBlocks:  publicSubnetCidrBlocks,
		})
		if err != nil {
			return err
		}

		ctx.Export("vpcId", output.Vpc.ID())

		return nil
	})
}

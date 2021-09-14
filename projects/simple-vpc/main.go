package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	vpcCidrBlock = "10.0.0.0/16"

	publicSubnetCidrBlocks  = []string{"10.0.0.0/24", "10.0.1.0/24", "10.0.2.0/24"}
	privateSubnetCidrBlocks = []string{"10.0.100.0/24", "10.0.101.0/24", "10.0.102.0/24"}
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		availabilityZones, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{})
		if err != nil {
			return err
		}

		// VPC
		vpc, err := ec2.NewVpc(ctx, "vpc", &ec2.VpcArgs{
			CidrBlock: pulumi.String(vpcCidrBlock),
			Tags:      pulumi.ToStringMap(namedTags(ctx, "vpc")),
		})
		if err != nil {
			return err
		}

		ctx.Export("vpcId", vpc.ID())

		publicSubnets, err := createSubnets(ctx, "public", vpc.ID(), publicSubnetCidrBlocks, availabilityZones.Names)
		if err != nil {
			return err
		}

		privateSubnets, err := createSubnets(ctx, "private", vpc.ID(), privateSubnetCidrBlocks, availabilityZones.Names)
		if err != nil {
			return err
		}

		iGW, err := createInternetGateway(ctx, vpc.ID())
		if err != nil {
			return fmt.Errorf("while creating InternetGateway: %v", err)
		}

		natGW, err := createNATGateway(ctx, publicSubnets[0].ID())
		if err != nil {
			return fmt.Errorf("while creating NAT Gateway: %v", err)
		}

		if _, err := createPublicRouteTable(ctx, vpc.ID(), publicSubnets, iGW); err != nil {
			return fmt.Errorf("while creating public route table: %v", err)
		}

		if _, err := createPrivateRouteTable(ctx, vpc.ID(), privateSubnets, natGW); err != nil {
			return fmt.Errorf("while creating private route table: %v", err)
		}

		return nil
	})
}

func createSubnets(ctx *pulumi.Context, prefix string, vpcID pulumi.StringInput, cidrBlocks []string, AZs []string) ([]*ec2.Subnet, error) {
	if len(cidrBlocks) != len(AZs) {
		return nil, fmt.Errorf("number of CIDR blocks and AZs is not equal")
	}

	subnets := make([]*ec2.Subnet, 0, len(cidrBlocks))

	for i, cidr := range cidrBlocks {
		az := AZs[i]

		name := fmt.Sprintf("%s-%d", prefix, i)
		subnet, err := ec2.NewSubnet(ctx, name, &ec2.SubnetArgs{
			VpcId:            vpcID,
			CidrBlock:        pulumi.String(cidr),
			AvailabilityZone: pulumi.StringPtr(az),
			Tags:             pulumi.ToStringMap(namedTags(ctx, name)),
		})
		if err != nil {
			return nil, err
		}

		subnets = append(subnets, subnet)
	}

	return subnets, nil
}

func createInternetGateway(ctx *pulumi.Context, vpcID pulumi.StringPtrInput) (*ec2.InternetGateway, error) {
	igw, err := ec2.NewInternetGateway(ctx, "igw", &ec2.InternetGatewayArgs{
		VpcId: vpcID,
		Tags:  pulumi.ToStringMap(namedTags(ctx, "igw")),
	})
	if err != nil {
		return nil, err
	}

	return igw, nil
}

func createNATGateway(ctx *pulumi.Context, subnetID pulumi.StringInput) (*ec2.NatGateway, error) {
	eip, err := ec2.NewEip(ctx, "nat-eip", &ec2.EipArgs{
		Vpc: pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	gw, err := ec2.NewNatGateway(ctx, "nat-gw", &ec2.NatGatewayArgs{
		SubnetId:     subnetID,
		AllocationId: eip.ID(),
		Tags:         pulumi.ToStringMap(commonTags(ctx)),
	})
	if err != nil {
		return nil, err
	}

	return gw, nil
}

func createPrivateRouteTable(ctx *pulumi.Context, vpcID pulumi.StringInput, subnets []*ec2.Subnet, natGW *ec2.NatGateway) (*ec2.RouteTable, error) {
	rt, err := ec2.NewRouteTable(ctx, "private-rt", &ec2.RouteTableArgs{
		VpcId: vpcID,
		Routes: ec2.RouteTableRouteArray{
			ec2.RouteTableRouteArgs{
				CidrBlock:    pulumi.String("0.0.0.0/0"),
				NatGatewayId: natGW.ID(),
			},
		},
		Tags: pulumi.ToStringMap(namedTags(ctx, "private-rt")),
	})
	if err != nil {
		return nil, err
	}

	for i, subnet := range subnets {
		name := fmt.Sprintf("private-subnet-assoc-%d", i)
		if _, err := ec2.NewRouteTableAssociation(ctx, name, &ec2.RouteTableAssociationArgs{
			RouteTableId: rt.ID(),
			SubnetId:     subnet.ID(),
		}); err != nil {
			return nil, err
		}
	}

	return rt, nil
}

func createPublicRouteTable(ctx *pulumi.Context, vpcID pulumi.StringInput, subnets []*ec2.Subnet, iGW *ec2.InternetGateway) (*ec2.RouteTable, error) {
	rt, err := ec2.NewRouteTable(ctx, "public-rt", &ec2.RouteTableArgs{
		VpcId: vpcID,
		Routes: ec2.RouteTableRouteArray{
			ec2.RouteTableRouteArgs{
				CidrBlock: pulumi.String("0.0.0.0/0"),
				GatewayId: iGW.ID(),
			},
		},
		Tags: pulumi.ToStringMap(namedTags(ctx, "public-rt")),
	})
	if err != nil {
		return nil, err
	}

	for i, subnet := range subnets {
		name := fmt.Sprintf("public-subnet-assoc-%d", i)
		if _, err := ec2.NewRouteTableAssociation(ctx, name, &ec2.RouteTableAssociationArgs{
			RouteTableId: rt.ID(),
			SubnetId:     subnet.ID(),
		}); err != nil {
			return nil, err
		}
	}

	return rt, nil
}

func namedTags(ctx *pulumi.Context, name string) map[string]string {
	tags := commonTags(ctx)
	tags["Name"] = fmt.Sprintf("%s-%s-%s", ctx.Project(), ctx.Stack(), name)
	return tags
}

func commonTags(ctx *pulumi.Context) map[string]string {
	return map[string]string{
		"Project":     ctx.Project(),
		"Environmen:": ctx.Stack(),
		"Name":        fmt.Sprintf("%s-%s", ctx.Project(), ctx.Stack()),
	}
}

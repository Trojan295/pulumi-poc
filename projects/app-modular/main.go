package main

import (
	"fmt"

	"github.com/Trojan295/pulumi-poc/pkg/ec2"
	"github.com/Trojan295/pulumi-poc/pkg/elb"
	awsec2 "github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	awselb "github.com/pulumi/pulumi-aws/sdk/v4/go/aws/elb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		name := fmt.Sprintf("%s-%s", ctx.Project(), ctx.Stack())

		vpcStackName := fmt.Sprintf("Trojan295/vpc-modular/%s", ctx.Stack())
		vpcStack, err := pulumi.NewStackReference(ctx, vpcStackName, nil)
		if err != nil {
			return err
		}

		vpcID := vpcStack.GetStringOutput(pulumi.String("vpcId"))

		publicSubnets := vpcStack.GetOutput(pulumi.String("publicSubnetIDs")).ApplyT(func(x interface{}) []string {
			y := x.([]interface{})
			r := make([]string, 0)
			for _, item := range y {
				r = append(r, item.(string))
			}
			return r
		}).(pulumi.StringArrayOutput)

		//privateSubnets := vpcStack.GetOutput(pulumi.String("privateSubnetIDs")).ApplyT(func(x interface{}) []string {
		//	y := x.([]interface{})
		//	r := make([]string, 0)
		//	for _, item := range y {
		//		r = append(r, item.(string))
		//	}
		//	return r
		//}).(pulumi.StringArrayOutput)

		ec2SgOutput, err := ec2.NewSecurityGroup(ctx, &ec2.SecurityGroupInput{
			Name:  name + "-ec2",
			VpcID: vpcID,
			Ingress: awsec2.SecurityGroupIngressArray{
				awsec2.SecurityGroupIngressArgs{
					Description: pulumi.String("http"),
					FromPort:    pulumi.Int(80),
					ToPort:      pulumi.Int(80),
					Protocol:    pulumi.String("tcp"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
				},
			},
			Egress: awsec2.SecurityGroupEgressArray{
				awsec2.SecurityGroupEgressArgs{
					Description: pulumi.String("all"),
					FromPort:    pulumi.Int(0),
					ToPort:      pulumi.Int(0),
					Protocol:    pulumi.String("ALL"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
				},
			},
		})

		elbSgOutput, err := ec2.NewSecurityGroup(ctx, &ec2.SecurityGroupInput{
			Name:  name + "-elb",
			VpcID: vpcID,
			Ingress: awsec2.SecurityGroupIngressArray{
				awsec2.SecurityGroupIngressArgs{
					Description: pulumi.String("http"),
					FromPort:    pulumi.Int(80),
					ToPort:      pulumi.Int(80),
					Protocol:    pulumi.String("tcp"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
				},
			},
			Egress: awsec2.SecurityGroupEgressArray{
				awsec2.SecurityGroupEgressArgs{
					Description: pulumi.String("all"),
					FromPort:    pulumi.Int(80),
					ToPort:      pulumi.Int(80),
					Protocol:    pulumi.String("TCP"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
				},
			},
		})

		elbOutput, err := elb.NewElb(ctx, &elb.ElbInput{
			Name:             name,
			SubnetIDs:        publicSubnets,
			SecurityGroupIDs: pulumi.StringArray{elbSgOutput.SecurityGroup.ID()},
			Listeners: awselb.LoadBalancerListenerArray{
				awselb.LoadBalancerListenerArgs{
					LbPort:           pulumi.Int(80),
					LbProtocol:       pulumi.String("http"),
					InstancePort:     pulumi.Int(80),
					InstanceProtocol: pulumi.String("http"),
				},
			},
		})
		if err != nil {
			return err
		}

		_, err = ec2.NewAsg(ctx, &ec2.AsgInput{
			Name:             name,
			AmiID:            pulumi.String("ami-0d1bf5b68307103c2"),
			InstanceType:     pulumi.String("t3a.micro"),
			UserData:         pulumi.String("IyEvYmluL2Jhc2ggLXhlCmFtYXpvbi1saW51eC1leHRyYXMgaW5zdGFsbCAteSBuZ2lueDEKClRPS0VOPWBjdXJsIC1YIFBVVCAiaHR0cDovLzE2OS4yNTQuMTY5LjI1NC9sYXRlc3QvYXBpL3Rva2VuIiAtSCAiWC1hd3MtZWMyLW1ldGFkYXRhLXRva2VuLXR0bC1zZWNvbmRzOiAyMTYwMCJgCgpjYXQgPDxFT0YgPiAvdXNyL3NoYXJlL25naW54L2h0bWwvaW5kZXguaHRtbApEYXRlOiAkKGRhdGUpCkFNSSBJRDogJChjdXJsIC1IICJYLWF3cy1lYzItbWV0YWRhdGEtdG9rZW46ICRUT0tFTiIgaHR0cDovLzE2OS4yNTQuMTY5LjI1NC9sYXRlc3QvbWV0YS1kYXRhL2FtaS1pZCkKRU9GCgpzeXN0ZW1jdGwgc3RhcnQgbmdpbngKc3lzdGVtY3RsIGVuYWJsZSBuZ2lueAoK"),
			SubnetIDs:        publicSubnets,
			LoadBalancerID:   elbOutput.LoadBlanacer.ID(),
			SecurityGroupIDs: pulumi.StringArray{ec2SgOutput.SecurityGroup.ID()},
		})
		if err != nil {
			return err
		}

		return nil
	})
}

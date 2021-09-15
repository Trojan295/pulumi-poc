package ec2

import (
	"github.com/Trojan295/pulumi-poc/pkg/utils"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SecurityGroupInput struct {
	Name    string
	VpcID   pulumi.StringPtrInput
	Ingress ec2.SecurityGroupIngressArrayInput
	Egress  ec2.SecurityGroupEgressArrayInput
}

type SecurityGroupOutput struct {
	SecurityGroup *ec2.SecurityGroup
}

func NewSecurityGroup(ctx *pulumi.Context, input *SecurityGroupInput) (*SecurityGroupOutput, error) {
	var (
		err    error
		output = &SecurityGroupOutput{}
	)

	output.SecurityGroup, err = ec2.NewSecurityGroup(ctx, input.Name, &ec2.SecurityGroupArgs{
		Name:        pulumi.String(input.Name),
		Description: pulumi.String(input.Name),
		Ingress:     input.Ingress,
		Egress:      input.Egress,
		VpcId:       input.VpcID,
		Tags:        pulumi.ToStringMap(utils.NewNamedTags(ctx, input.Name)),
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

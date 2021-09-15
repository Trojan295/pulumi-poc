package elb

import (
	"github.com/Trojan295/pulumi-poc/pkg/utils"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/elb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ElbInput struct {
	Name             string
	SubnetIDs        pulumi.StringArrayInput
	Listeners        elb.LoadBalancerListenerArray
	SecurityGroupIDs pulumi.StringArrayInput
}

type ElbOutput struct {
	LoadBlanacer *elb.LoadBalancer
}

func NewElb(ctx *pulumi.Context, input *ElbInput) (*ElbOutput, error) {
	var (
		err    error
		output = &ElbOutput{}
	)

	output.LoadBlanacer, err = elb.NewLoadBalancer(ctx, "elb", &elb.LoadBalancerArgs{
		Subnets:        input.SubnetIDs,
		Listeners:      input.Listeners,
		SecurityGroups: input.SecurityGroupIDs,
		Tags:           pulumi.ToStringMap(utils.NewNamedTags(ctx, input.Name)),
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

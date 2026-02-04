// Package ec2 provides AWS EC2 operations
package ec2

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awscommon "github.com/inovacc/omni/internal/cli/aws"
)

// Client wraps the EC2 client
type Client struct {
	client  *ec2.Client
	printer *awscommon.Printer
}

// NewClient creates a new EC2 client
func NewClient(cfg aws.Config, w io.Writer, format awscommon.OutputFormat) *Client {
	return &Client{
		client:  ec2.NewFromConfig(cfg),
		printer: awscommon.NewPrinter(w, format),
	}
}

// Instance represents an EC2 instance
type Instance struct {
	InstanceId       string            `json:"InstanceId"`
	InstanceType     string            `json:"InstanceType"`
	State            string            `json:"State"`
	PublicIpAddress  string            `json:"PublicIpAddress,omitempty"`
	PrivateIpAddress string            `json:"PrivateIpAddress,omitempty"`
	LaunchTime       time.Time         `json:"LaunchTime,omitempty"`
	Tags             map[string]string `json:"Tags,omitempty"`
	VpcId            string            `json:"VpcId,omitempty"`
	SubnetId         string            `json:"SubnetId,omitempty"`
	ImageId          string            `json:"ImageId,omitempty"`
	KeyName          string            `json:"KeyName,omitempty"`
	SecurityGroups   []string          `json:"SecurityGroups,omitempty"`
}

// DescribeInstancesInput contains parameters for DescribeInstances
type DescribeInstancesInput struct {
	InstanceIds []string
	Filters     []Filter
	MaxResults  int32
}

// Filter represents a filter for EC2 queries
type Filter struct {
	Name   string
	Values []string
}

// DescribeInstances describes EC2 instances
func (c *Client) DescribeInstances(ctx context.Context, input DescribeInstancesInput) ([]Instance, error) {
	params := &ec2.DescribeInstancesInput{}

	if len(input.InstanceIds) > 0 {
		params.InstanceIds = input.InstanceIds
	}

	if len(input.Filters) > 0 {
		filters := make([]types.Filter, 0, len(input.Filters))
		for _, f := range input.Filters {
			filters = append(filters, types.Filter{
				Name:   aws.String(f.Name),
				Values: f.Values,
			})
		}
		params.Filters = filters
	}

	if input.MaxResults > 0 {
		params.MaxResults = aws.Int32(input.MaxResults)
	}

	var instances []Instance

	paginator := ec2.NewDescribeInstancesPaginator(c.client, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describe-instances: %w", err)
		}

		for _, reservation := range page.Reservations {
			for _, inst := range reservation.Instances {
				instances = append(instances, instanceFromAWS(&inst))
			}
		}
	}

	return instances, nil
}

// PrintInstances prints instances
func (c *Client) PrintInstances(instances []Instance) error {
	return c.printer.PrintJSON(map[string]any{
		"Reservations": []map[string]any{
			{"Instances": instances},
		},
	})
}

// StartInstances starts EC2 instances
func (c *Client) StartInstances(ctx context.Context, instanceIds []string) ([]InstanceStateChange, error) {
	result, err := c.client.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return nil, fmt.Errorf("start-instances: %w", err)
	}

	var changes []InstanceStateChange
	for _, change := range result.StartingInstances {
		changes = append(changes, InstanceStateChange{
			InstanceId:    aws.ToString(change.InstanceId),
			CurrentState:  string(change.CurrentState.Name),
			PreviousState: string(change.PreviousState.Name),
		})
	}

	return changes, nil
}

// StopInstances stops EC2 instances
func (c *Client) StopInstances(ctx context.Context, instanceIds []string, force bool) ([]InstanceStateChange, error) {
	result, err := c.client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: instanceIds,
		Force:       aws.Bool(force),
	})
	if err != nil {
		return nil, fmt.Errorf("stop-instances: %w", err)
	}

	var changes []InstanceStateChange
	for _, change := range result.StoppingInstances {
		changes = append(changes, InstanceStateChange{
			InstanceId:    aws.ToString(change.InstanceId),
			CurrentState:  string(change.CurrentState.Name),
			PreviousState: string(change.PreviousState.Name),
		})
	}

	return changes, nil
}

// InstanceStateChange represents an instance state change
type InstanceStateChange struct {
	InstanceId    string `json:"InstanceId"`
	CurrentState  string `json:"CurrentState"`
	PreviousState string `json:"PreviousState"`
}

// PrintStateChanges prints instance state changes
func (c *Client) PrintStateChanges(changes []InstanceStateChange) error {
	return c.printer.PrintJSON(map[string]any{
		"StartingInstances": changes,
	})
}

// VPC represents a VPC
type VPC struct {
	VpcId           string            `json:"VpcId"`
	CidrBlock       string            `json:"CidrBlock"`
	State           string            `json:"State"`
	IsDefault       bool              `json:"IsDefault"`
	Tags            map[string]string `json:"Tags,omitempty"`
	DhcpOptionsId   string            `json:"DhcpOptionsId,omitempty"`
	InstanceTenancy string            `json:"InstanceTenancy,omitempty"`
}

// DescribeVpcs describes VPCs
func (c *Client) DescribeVpcs(ctx context.Context, vpcIds []string, filters []Filter) ([]VPC, error) {
	params := &ec2.DescribeVpcsInput{}

	if len(vpcIds) > 0 {
		params.VpcIds = vpcIds
	}

	if len(filters) > 0 {
		ec2Filters := make([]types.Filter, 0, len(filters))
		for _, f := range filters {
			ec2Filters = append(ec2Filters, types.Filter{
				Name:   aws.String(f.Name),
				Values: f.Values,
			})
		}
		params.Filters = ec2Filters
	}

	result, err := c.client.DescribeVpcs(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("describe-vpcs: %w", err)
	}

	var vpcs []VPC
	for _, v := range result.Vpcs {
		vpcs = append(vpcs, VPC{
			VpcId:           aws.ToString(v.VpcId),
			CidrBlock:       aws.ToString(v.CidrBlock),
			State:           string(v.State),
			IsDefault:       aws.ToBool(v.IsDefault),
			Tags:            tagsToMap(v.Tags),
			DhcpOptionsId:   aws.ToString(v.DhcpOptionsId),
			InstanceTenancy: string(v.InstanceTenancy),
		})
	}

	return vpcs, nil
}

// PrintVpcs prints VPCs
func (c *Client) PrintVpcs(vpcs []VPC) error {
	return c.printer.PrintJSON(map[string]any{
		"Vpcs": vpcs,
	})
}

// SecurityGroup represents a security group
type SecurityGroup struct {
	GroupId           string            `json:"GroupId"`
	GroupName         string            `json:"GroupName"`
	Description       string            `json:"Description"`
	VpcId             string            `json:"VpcId,omitempty"`
	Tags              map[string]string `json:"Tags,omitempty"`
	IpPermissions     []IpPermission    `json:"IpPermissions,omitempty"`
	IpPermissionsEgress []IpPermission  `json:"IpPermissionsEgress,omitempty"`
}

// IpPermission represents an IP permission rule
type IpPermission struct {
	IpProtocol string     `json:"IpProtocol"`
	FromPort   int32      `json:"FromPort,omitempty"`
	ToPort     int32      `json:"ToPort,omitempty"`
	IpRanges   []IpRange  `json:"IpRanges,omitempty"`
}

// IpRange represents an IP range
type IpRange struct {
	CidrIp      string `json:"CidrIp"`
	Description string `json:"Description,omitempty"`
}

// DescribeSecurityGroups describes security groups
func (c *Client) DescribeSecurityGroups(ctx context.Context, groupIds []string, filters []Filter) ([]SecurityGroup, error) {
	params := &ec2.DescribeSecurityGroupsInput{}

	if len(groupIds) > 0 {
		params.GroupIds = groupIds
	}

	if len(filters) > 0 {
		ec2Filters := make([]types.Filter, 0, len(filters))
		for _, f := range filters {
			ec2Filters = append(ec2Filters, types.Filter{
				Name:   aws.String(f.Name),
				Values: f.Values,
			})
		}
		params.Filters = ec2Filters
	}

	result, err := c.client.DescribeSecurityGroups(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("describe-security-groups: %w", err)
	}

	var groups []SecurityGroup
	for _, sg := range result.SecurityGroups {
		group := SecurityGroup{
			GroupId:     aws.ToString(sg.GroupId),
			GroupName:   aws.ToString(sg.GroupName),
			Description: aws.ToString(sg.Description),
			VpcId:       aws.ToString(sg.VpcId),
			Tags:        tagsToMap(sg.Tags),
		}

		// Convert IP permissions
		for _, perm := range sg.IpPermissions {
			ipPerm := IpPermission{
				IpProtocol: aws.ToString(perm.IpProtocol),
				FromPort:   aws.ToInt32(perm.FromPort),
				ToPort:     aws.ToInt32(perm.ToPort),
			}
			for _, ipr := range perm.IpRanges {
				ipPerm.IpRanges = append(ipPerm.IpRanges, IpRange{
					CidrIp:      aws.ToString(ipr.CidrIp),
					Description: aws.ToString(ipr.Description),
				})
			}
			group.IpPermissions = append(group.IpPermissions, ipPerm)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// PrintSecurityGroups prints security groups
func (c *Client) PrintSecurityGroups(groups []SecurityGroup) error {
	return c.printer.PrintJSON(map[string]any{
		"SecurityGroups": groups,
	})
}

// instanceFromAWS converts an AWS instance to our type
func instanceFromAWS(inst *types.Instance) Instance {
	instance := Instance{
		InstanceId:       aws.ToString(inst.InstanceId),
		InstanceType:     string(inst.InstanceType),
		PrivateIpAddress: aws.ToString(inst.PrivateIpAddress),
		PublicIpAddress:  aws.ToString(inst.PublicIpAddress),
		VpcId:            aws.ToString(inst.VpcId),
		SubnetId:         aws.ToString(inst.SubnetId),
		ImageId:          aws.ToString(inst.ImageId),
		KeyName:          aws.ToString(inst.KeyName),
		Tags:             tagsToMap(inst.Tags),
	}

	if inst.State != nil {
		instance.State = string(inst.State.Name)
	}

	if inst.LaunchTime != nil {
		instance.LaunchTime = *inst.LaunchTime
	}

	for _, sg := range inst.SecurityGroups {
		instance.SecurityGroups = append(instance.SecurityGroups, aws.ToString(sg.GroupId))
	}

	return instance
}

// tagsToMap converts AWS tags to a map
func tagsToMap(tags []types.Tag) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	m := make(map[string]string, len(tags))
	for _, t := range tags {
		m[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}
	return m
}

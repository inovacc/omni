// Package iam provides AWS IAM operations
package iam

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	awscommon "github.com/inovacc/omni/internal/cli/aws"
)

// Client wraps the IAM client
type Client struct {
	client  *iam.Client
	printer *awscommon.Printer
}

// NewClient creates a new IAM client
func NewClient(cfg aws.Config, w io.Writer, format awscommon.OutputFormat, endpointURL string) *Client {
	var client *iam.Client
	if endpointURL != "" {
		client = iam.NewFromConfig(cfg, func(o *iam.Options) {
			o.BaseEndpoint = aws.String(endpointURL)
		})
	} else {
		client = iam.NewFromConfig(cfg)
	}

	return &Client{
		client:  client,
		printer: awscommon.NewPrinter(w, format),
	}
}

// User represents an IAM user
type User struct {
	UserName         string    `json:"UserName"`
	UserId           string    `json:"UserId"`
	Arn              string    `json:"Arn"`
	Path             string    `json:"Path,omitempty"`
	CreateDate       time.Time `json:"CreateDate,omitzero"`
	PasswordLastUsed time.Time `json:"PasswordLastUsed,omitzero"`
	Tags             []Tag     `json:"Tags,omitempty"`
}

// Tag represents an IAM tag
type Tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// GetUser retrieves information about an IAM user
func (c *Client) GetUser(ctx context.Context, userName string) (*User, error) {
	params := &iam.GetUserInput{}
	if userName != "" {
		params.UserName = aws.String(userName)
	}

	result, err := c.client.GetUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get-user: %w", err)
	}

	user := &User{
		UserName: aws.ToString(result.User.UserName),
		UserId:   aws.ToString(result.User.UserId),
		Arn:      aws.ToString(result.User.Arn),
		Path:     aws.ToString(result.User.Path),
	}

	if result.User.CreateDate != nil {
		user.CreateDate = *result.User.CreateDate
	}

	if result.User.PasswordLastUsed != nil {
		user.PasswordLastUsed = *result.User.PasswordLastUsed
	}

	for _, t := range result.User.Tags {
		user.Tags = append(user.Tags, Tag{
			Key:   aws.ToString(t.Key),
			Value: aws.ToString(t.Value),
		})
	}

	return user, nil
}

// PrintUser prints a user
func (c *Client) PrintUser(user *User) error {
	return c.printer.PrintJSON(map[string]any{
		"User": user,
	})
}

// Role represents an IAM role
type Role struct {
	RoleName                 string    `json:"RoleName"`
	RoleId                   string    `json:"RoleId"`
	Arn                      string    `json:"Arn"`
	Path                     string    `json:"Path,omitempty"`
	CreateDate               time.Time `json:"CreateDate,omitzero"`
	AssumeRolePolicyDocument string    `json:"AssumeRolePolicyDocument,omitempty"`
	Description              string    `json:"Description,omitempty"`
	MaxSessionDuration       int32     `json:"MaxSessionDuration,omitempty"`
	Tags                     []Tag     `json:"Tags,omitempty"`
}

// GetRole retrieves information about an IAM role
func (c *Client) GetRole(ctx context.Context, roleName string) (*Role, error) {
	result, err := c.client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return nil, fmt.Errorf("get-role: %w", err)
	}

	role := &Role{
		RoleName:                 aws.ToString(result.Role.RoleName),
		RoleId:                   aws.ToString(result.Role.RoleId),
		Arn:                      aws.ToString(result.Role.Arn),
		Path:                     aws.ToString(result.Role.Path),
		AssumeRolePolicyDocument: aws.ToString(result.Role.AssumeRolePolicyDocument),
		Description:              aws.ToString(result.Role.Description),
		MaxSessionDuration:       aws.ToInt32(result.Role.MaxSessionDuration),
	}

	if result.Role.CreateDate != nil {
		role.CreateDate = *result.Role.CreateDate
	}

	for _, t := range result.Role.Tags {
		role.Tags = append(role.Tags, Tag{
			Key:   aws.ToString(t.Key),
			Value: aws.ToString(t.Value),
		})
	}

	return role, nil
}

// PrintRole prints a role
func (c *Client) PrintRole(role *Role) error {
	return c.printer.PrintJSON(map[string]any{
		"Role": role,
	})
}

// ListRolesInput contains parameters for ListRoles
type ListRolesInput struct {
	PathPrefix string
	MaxItems   int32
}

// ListRoles lists IAM roles
func (c *Client) ListRoles(ctx context.Context, input ListRolesInput) ([]Role, error) {
	params := &iam.ListRolesInput{}

	if input.PathPrefix != "" {
		params.PathPrefix = aws.String(input.PathPrefix)
	}

	if input.MaxItems > 0 {
		params.MaxItems = aws.Int32(input.MaxItems)
	}

	var roles []Role

	paginator := iam.NewListRolesPaginator(c.client, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list-roles: %w", err)
		}

		for _, r := range page.Roles {
			role := Role{
				RoleName:                 aws.ToString(r.RoleName),
				RoleId:                   aws.ToString(r.RoleId),
				Arn:                      aws.ToString(r.Arn),
				Path:                     aws.ToString(r.Path),
				AssumeRolePolicyDocument: aws.ToString(r.AssumeRolePolicyDocument),
				Description:              aws.ToString(r.Description),
				MaxSessionDuration:       aws.ToInt32(r.MaxSessionDuration),
			}
			if r.CreateDate != nil {
				role.CreateDate = *r.CreateDate
			}

			roles = append(roles, role)
		}
	}

	return roles, nil
}

// PrintRoles prints roles
func (c *Client) PrintRoles(roles []Role) error {
	return c.printer.PrintJSON(map[string]any{
		"Roles": roles,
	})
}

// Policy represents an IAM policy
type Policy struct {
	PolicyName                    string    `json:"PolicyName"`
	PolicyId                      string    `json:"PolicyId"`
	Arn                           string    `json:"Arn"`
	Path                          string    `json:"Path,omitempty"`
	DefaultVersionId              string    `json:"DefaultVersionId,omitempty"`
	AttachmentCount               int32     `json:"AttachmentCount,omitempty"`
	PermissionsBoundaryUsageCount int32     `json:"PermissionsBoundaryUsageCount,omitempty"`
	IsAttachable                  bool      `json:"IsAttachable"`
	CreateDate                    time.Time `json:"CreateDate,omitzero"`
	UpdateDate                    time.Time `json:"UpdateDate,omitzero"`
	Description                   string    `json:"Description,omitempty"`
	Tags                          []Tag     `json:"Tags,omitempty"`
}

// ListPoliciesInput contains parameters for ListPolicies
type ListPoliciesInput struct {
	Scope             string // All, AWS, Local
	OnlyAttached      bool
	PathPrefix        string
	PolicyUsageFilter string // PermissionsPolicy, PermissionsBoundary
	MaxItems          int32
}

// ListPolicies lists IAM policies
func (c *Client) ListPolicies(ctx context.Context, input ListPoliciesInput) ([]Policy, error) {
	params := &iam.ListPoliciesInput{}

	if input.Scope != "" {
		params.Scope = iamPolicyScope(input.Scope)
	}

	if input.OnlyAttached {
		params.OnlyAttached = input.OnlyAttached
	}

	if input.PathPrefix != "" {
		params.PathPrefix = aws.String(input.PathPrefix)
	}

	if input.MaxItems > 0 {
		params.MaxItems = aws.Int32(input.MaxItems)
	}

	var policies []Policy

	paginator := iam.NewListPoliciesPaginator(c.client, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list-policies: %w", err)
		}

		for _, p := range page.Policies {
			policy := Policy{
				PolicyName:                    aws.ToString(p.PolicyName),
				PolicyId:                      aws.ToString(p.PolicyId),
				Arn:                           aws.ToString(p.Arn),
				Path:                          aws.ToString(p.Path),
				DefaultVersionId:              aws.ToString(p.DefaultVersionId),
				AttachmentCount:               aws.ToInt32(p.AttachmentCount),
				PermissionsBoundaryUsageCount: aws.ToInt32(p.PermissionsBoundaryUsageCount),
				IsAttachable:                  p.IsAttachable,
				Description:                   aws.ToString(p.Description),
			}
			if p.CreateDate != nil {
				policy.CreateDate = *p.CreateDate
			}

			if p.UpdateDate != nil {
				policy.UpdateDate = *p.UpdateDate
			}

			policies = append(policies, policy)
		}
	}

	return policies, nil
}

// PrintPolicies prints policies
func (c *Client) PrintPolicies(policies []Policy) error {
	return c.printer.PrintJSON(map[string]any{
		"Policies": policies,
	})
}

// GetPolicy retrieves information about an IAM policy
func (c *Client) GetPolicy(ctx context.Context, policyArn string) (*Policy, error) {
	result, err := c.client.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: aws.String(policyArn),
	})
	if err != nil {
		return nil, fmt.Errorf("get-policy: %w", err)
	}

	policy := &Policy{
		PolicyName:                    aws.ToString(result.Policy.PolicyName),
		PolicyId:                      aws.ToString(result.Policy.PolicyId),
		Arn:                           aws.ToString(result.Policy.Arn),
		Path:                          aws.ToString(result.Policy.Path),
		DefaultVersionId:              aws.ToString(result.Policy.DefaultVersionId),
		AttachmentCount:               aws.ToInt32(result.Policy.AttachmentCount),
		PermissionsBoundaryUsageCount: aws.ToInt32(result.Policy.PermissionsBoundaryUsageCount),
		IsAttachable:                  result.Policy.IsAttachable,
		Description:                   aws.ToString(result.Policy.Description),
	}

	if result.Policy.CreateDate != nil {
		policy.CreateDate = *result.Policy.CreateDate
	}

	if result.Policy.UpdateDate != nil {
		policy.UpdateDate = *result.Policy.UpdateDate
	}

	for _, t := range result.Policy.Tags {
		policy.Tags = append(policy.Tags, Tag{
			Key:   aws.ToString(t.Key),
			Value: aws.ToString(t.Value),
		})
	}

	return policy, nil
}

// PrintPolicy prints a policy
func (c *Client) PrintPolicy(policy *Policy) error {
	return c.printer.PrintJSON(map[string]any{
		"Policy": policy,
	})
}

// iamPolicyScope converts a string to IAM policy scope type
func iamPolicyScope(s string) types.PolicyScopeType {
	switch s {
	case "AWS":
		return types.PolicyScopeTypeAws
	case "Local":
		return types.PolicyScopeTypeLocal
	default:
		return types.PolicyScopeTypeAll
	}
}

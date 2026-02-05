// Package sts provides AWS STS operations
package sts

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	awscommon "github.com/inovacc/omni/internal/cli/aws"
)

// Client wraps the STS client
type Client struct {
	client  *sts.Client
	printer *awscommon.Printer
}

// NewClient creates a new STS client
func NewClient(cfg aws.Config, w io.Writer, format awscommon.OutputFormat, endpointURL string) *Client {
	var client *sts.Client
	if endpointURL != "" {
		client = sts.NewFromConfig(cfg, func(o *sts.Options) {
			o.BaseEndpoint = aws.String(endpointURL)
		})
	} else {
		client = sts.NewFromConfig(cfg)
	}

	return &Client{
		client:  client,
		printer: awscommon.NewPrinter(w, format),
	}
}

// CallerIdentity represents the caller identity
type CallerIdentity struct {
	UserId  string `json:"UserId"`
	Account string `json:"Account"`
	Arn     string `json:"Arn"`
}

// GetCallerIdentity returns details about the IAM identity calling the API
func (c *Client) GetCallerIdentity(ctx context.Context) (*CallerIdentity, error) {
	result, err := c.client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("get-caller-identity: %w", err)
	}

	return &CallerIdentity{
		UserId:  aws.ToString(result.UserId),
		Account: aws.ToString(result.Account),
		Arn:     aws.ToString(result.Arn),
	}, nil
}

// PrintCallerIdentity prints the caller identity
func (c *Client) PrintCallerIdentity(identity *CallerIdentity) error {
	return c.printer.PrintJSON(identity)
}

// AssumeRoleInput contains parameters for AssumeRole
type AssumeRoleInput struct {
	RoleArn         string
	RoleSessionName string
	DurationSeconds int32
	ExternalId      string
	Policy          string
	SerialNumber    string
	TokenCode       string
}

// Credentials represents temporary security credentials
type Credentials struct {
	AccessKeyId     string    `json:"AccessKeyId"`
	SecretAccessKey string    `json:"SecretAccessKey"`
	SessionToken    string    `json:"SessionToken"`
	Expiration      time.Time `json:"Expiration"`
}

// AssumeRoleOutput represents the result of AssumeRole
type AssumeRoleOutput struct {
	Credentials      Credentials     `json:"Credentials"`
	AssumedRoleUser  AssumedRoleUser `json:"AssumedRoleUser"`
	PackedPolicySize *int32          `json:"PackedPolicySize,omitempty"`
}

// AssumedRoleUser represents the assumed role user
type AssumedRoleUser struct {
	AssumedRoleId string `json:"AssumedRoleId"`
	Arn           string `json:"Arn"`
}

// AssumeRole assumes an IAM role
func (c *Client) AssumeRole(ctx context.Context, input AssumeRoleInput) (*AssumeRoleOutput, error) {
	params := &sts.AssumeRoleInput{
		RoleArn:         aws.String(input.RoleArn),
		RoleSessionName: aws.String(input.RoleSessionName),
	}

	if input.DurationSeconds > 0 {
		params.DurationSeconds = aws.Int32(input.DurationSeconds)
	}

	if input.ExternalId != "" {
		params.ExternalId = aws.String(input.ExternalId)
	}

	if input.Policy != "" {
		params.Policy = aws.String(input.Policy)
	}

	if input.SerialNumber != "" && input.TokenCode != "" {
		params.SerialNumber = aws.String(input.SerialNumber)
		params.TokenCode = aws.String(input.TokenCode)
	}

	result, err := c.client.AssumeRole(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("assume-role: %w", err)
	}

	output := &AssumeRoleOutput{
		Credentials: Credentials{
			AccessKeyId:     aws.ToString(result.Credentials.AccessKeyId),
			SecretAccessKey: aws.ToString(result.Credentials.SecretAccessKey),
			SessionToken:    aws.ToString(result.Credentials.SessionToken),
			Expiration:      aws.ToTime(result.Credentials.Expiration),
		},
		AssumedRoleUser: AssumedRoleUser{
			AssumedRoleId: aws.ToString(result.AssumedRoleUser.AssumedRoleId),
			Arn:           aws.ToString(result.AssumedRoleUser.Arn),
		},
	}

	if result.PackedPolicySize != nil {
		output.PackedPolicySize = result.PackedPolicySize
	}

	return output, nil
}

// PrintAssumeRole prints the assume role output
func (c *Client) PrintAssumeRole(output *AssumeRoleOutput) error {
	return c.printer.PrintJSON(output)
}

// GetSessionToken gets temporary credentials
func (c *Client) GetSessionToken(ctx context.Context, durationSeconds int32, serialNumber, tokenCode string) (*Credentials, error) {
	params := &sts.GetSessionTokenInput{}

	if durationSeconds > 0 {
		params.DurationSeconds = aws.Int32(durationSeconds)
	}

	if serialNumber != "" && tokenCode != "" {
		params.SerialNumber = aws.String(serialNumber)
		params.TokenCode = aws.String(tokenCode)
	}

	result, err := c.client.GetSessionToken(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get-session-token: %w", err)
	}

	return &Credentials{
		AccessKeyId:     aws.ToString(result.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(result.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(result.Credentials.SessionToken),
		Expiration:      aws.ToTime(result.Credentials.Expiration),
	}, nil
}

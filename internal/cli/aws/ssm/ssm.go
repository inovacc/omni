// Package ssm provides AWS SSM Parameter Store operations
package ssm

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	awscommon "github.com/inovacc/omni/internal/cli/aws"
)

// Client wraps the SSM client
type Client struct {
	client  *ssm.Client
	printer *awscommon.Printer
}

// NewClient creates a new SSM client
func NewClient(cfg aws.Config, w io.Writer, format awscommon.OutputFormat) *Client {
	return &Client{
		client:  ssm.NewFromConfig(cfg),
		printer: awscommon.NewPrinter(w, format),
	}
}

// Parameter represents an SSM parameter
type Parameter struct {
	Name             string    `json:"Name"`
	Type             string    `json:"Type"`
	Value            string    `json:"Value"`
	Version          int64     `json:"Version"`
	LastModifiedDate time.Time `json:"LastModifiedDate,omitempty"`
	ARN              string    `json:"ARN,omitempty"`
	DataType         string    `json:"DataType,omitempty"`
}

// GetParameterInput contains parameters for GetParameter
type GetParameterInput struct {
	Name           string
	WithDecryption bool
}

// GetParameter retrieves a parameter by name
func (c *Client) GetParameter(ctx context.Context, input GetParameterInput) (*Parameter, error) {
	result, err := c.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(input.Name),
		WithDecryption: aws.Bool(input.WithDecryption),
	})
	if err != nil {
		return nil, fmt.Errorf("get-parameter: %w", err)
	}

	return parameterFromAWS(result.Parameter), nil
}

// PrintParameter prints a parameter
func (c *Client) PrintParameter(param *Parameter) error {
	return c.printer.PrintJSON(map[string]any{
		"Parameter": param,
	})
}

// GetParametersInput contains parameters for GetParameters
type GetParametersInput struct {
	Names          []string
	WithDecryption bool
}

// GetParametersOutput represents the result of GetParameters
type GetParametersOutput struct {
	Parameters        []Parameter `json:"Parameters"`
	InvalidParameters []string    `json:"InvalidParameters,omitempty"`
}

// GetParameters retrieves multiple parameters by name
func (c *Client) GetParameters(ctx context.Context, input GetParametersInput) (*GetParametersOutput, error) {
	result, err := c.client.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          input.Names,
		WithDecryption: aws.Bool(input.WithDecryption),
	})
	if err != nil {
		return nil, fmt.Errorf("get-parameters: %w", err)
	}

	output := &GetParametersOutput{
		Parameters:        make([]Parameter, 0, len(result.Parameters)),
		InvalidParameters: result.InvalidParameters,
	}

	for _, p := range result.Parameters {
		output.Parameters = append(output.Parameters, *parameterFromAWS(&p))
	}

	return output, nil
}

// GetParametersByPathInput contains parameters for GetParametersByPath
type GetParametersByPathInput struct {
	Path           string
	WithDecryption bool
	Recursive      bool
	MaxResults     int32
}

// GetParametersByPath retrieves parameters by path hierarchy
func (c *Client) GetParametersByPath(ctx context.Context, input GetParametersByPathInput) ([]Parameter, error) {
	var parameters []Parameter

	paginator := ssm.NewGetParametersByPathPaginator(c.client, &ssm.GetParametersByPathInput{
		Path:           aws.String(input.Path),
		WithDecryption: aws.Bool(input.WithDecryption),
		Recursive:      aws.Bool(input.Recursive),
		MaxResults:     aws.Int32(input.MaxResults),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("get-parameters-by-path: %w", err)
		}

		for _, p := range page.Parameters {
			parameters = append(parameters, *parameterFromAWS(&p))
		}
	}

	return parameters, nil
}

// PrintParameters prints parameters
func (c *Client) PrintParameters(params []Parameter) error {
	return c.printer.PrintJSON(map[string]any{
		"Parameters": params,
	})
}

// PutParameterInput contains parameters for PutParameter
type PutParameterInput struct {
	Name        string
	Value       string
	Type        string // String, StringList, SecureString
	Description string
	Overwrite   bool
	KeyId       string // KMS key for SecureString
	Tags        map[string]string
}

// PutParameterOutput represents the result of PutParameter
type PutParameterOutput struct {
	Version int64  `json:"Version"`
	Tier    string `json:"Tier"`
}

// PutParameter creates or updates a parameter
func (c *Client) PutParameter(ctx context.Context, input PutParameterInput) (*PutParameterOutput, error) {
	params := &ssm.PutParameterInput{
		Name:      aws.String(input.Name),
		Value:     aws.String(input.Value),
		Overwrite: aws.Bool(input.Overwrite),
	}

	// Set type
	switch input.Type {
	case "String":
		params.Type = types.ParameterTypeString
	case "StringList":
		params.Type = types.ParameterTypeStringList
	case "SecureString":
		params.Type = types.ParameterTypeSecureString
		if input.KeyId != "" {
			params.KeyId = aws.String(input.KeyId)
		}
	default:
		params.Type = types.ParameterTypeString
	}

	if input.Description != "" {
		params.Description = aws.String(input.Description)
	}

	if len(input.Tags) > 0 {
		tags := make([]types.Tag, 0, len(input.Tags))
		for k, v := range input.Tags {
			tags = append(tags, types.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
		params.Tags = tags
	}

	result, err := c.client.PutParameter(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("put-parameter: %w", err)
	}

	return &PutParameterOutput{
		Version: result.Version,
		Tier:    string(result.Tier),
	}, nil
}

// PrintPutParameter prints put parameter output
func (c *Client) PrintPutParameter(output *PutParameterOutput) error {
	return c.printer.PrintJSON(output)
}

// DeleteParameter deletes a parameter
func (c *Client) DeleteParameter(ctx context.Context, name string) error {
	_, err := c.client.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("delete-parameter: %w", err)
	}
	return nil
}

// parameterFromAWS converts an AWS parameter to our type
func parameterFromAWS(p *types.Parameter) *Parameter {
	if p == nil {
		return nil
	}

	param := &Parameter{
		Name:     aws.ToString(p.Name),
		Type:     string(p.Type),
		Value:    aws.ToString(p.Value),
		Version:  p.Version,
		DataType: aws.ToString(p.DataType),
		ARN:      aws.ToString(p.ARN),
	}

	if p.LastModifiedDate != nil {
		param.LastModifiedDate = *p.LastModifiedDate
	}

	return param
}

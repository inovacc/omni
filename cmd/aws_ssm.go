package cmd

import (
	"context"
	"fmt"

	awscommon "github.com/inovacc/omni/internal/cli/aws"
	"github.com/inovacc/omni/internal/cli/aws/ssm"
	"github.com/spf13/cobra"
)

var ssmCmd = &cobra.Command{
	Use:   "ssm",
	Short: "AWS SSM Parameter Store operations",
	Long:  `AWS Systems Manager Parameter Store operations.`,
}

var ssmGetParameterCmd = &cobra.Command{
	Use:   "get-parameter",
	Short: "Get a parameter value",
	Long: `Retrieves information about a parameter.

Examples:
  omni aws ssm get-parameter --name /app/config
  omni aws ssm get-parameter --name /app/secret --with-decryption`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")
		endpointURL, _ := cmd.Flags().GetString("endpoint-url")

		name, _ := cmd.Flags().GetString("name")
		withDecryption, _ := cmd.Flags().GetBool("with-decryption")

		opts := awscommon.Options{
			Profile:     profile,
			Region:      region,
			Output:      output,
			EndpointURL: endpointURL,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := ssm.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output), awscommon.GetEndpointURL(opts))

		param, err := client.GetParameter(ctx, ssm.GetParameterInput{
			Name:           name,
			WithDecryption: withDecryption,
		})
		if err != nil {
			return err
		}

		return client.PrintParameter(param)
	},
}

var ssmGetParametersCmd = &cobra.Command{
	Use:   "get-parameters",
	Short: "Get multiple parameters",
	Long: `Retrieves information about multiple parameters.

Examples:
  omni aws ssm get-parameters --names /app/config,/app/secret
  omni aws ssm get-parameters --names /app/config --names /app/secret --with-decryption`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")
		endpointURL, _ := cmd.Flags().GetString("endpoint-url")

		names, _ := cmd.Flags().GetStringSlice("names")
		withDecryption, _ := cmd.Flags().GetBool("with-decryption")

		opts := awscommon.Options{
			Profile:     profile,
			Region:      region,
			Output:      output,
			EndpointURL: endpointURL,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := ssm.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output), awscommon.GetEndpointURL(opts))

		result, err := client.GetParameters(ctx, ssm.GetParametersInput{
			Names:          names,
			WithDecryption: withDecryption,
		})
		if err != nil {
			return err
		}

		return awscommon.PrintJSON(cmd.OutOrStdout(), result)
	},
}

var ssmGetParametersByPathCmd = &cobra.Command{
	Use:   "get-parameters-by-path",
	Short: "Get parameters by path",
	Long: `Retrieves all parameters within a hierarchy.

Examples:
  omni aws ssm get-parameters-by-path --path /app/
  omni aws ssm get-parameters-by-path --path /app/ --recursive --with-decryption`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")
		endpointURL, _ := cmd.Flags().GetString("endpoint-url")

		path, _ := cmd.Flags().GetString("path")
		withDecryption, _ := cmd.Flags().GetBool("with-decryption")
		recursive, _ := cmd.Flags().GetBool("recursive")
		maxResults, _ := cmd.Flags().GetInt32("max-results")

		opts := awscommon.Options{
			Profile:     profile,
			Region:      region,
			Output:      output,
			EndpointURL: endpointURL,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := ssm.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output), awscommon.GetEndpointURL(opts))

		params, err := client.GetParametersByPath(ctx, ssm.GetParametersByPathInput{
			Path:           path,
			WithDecryption: withDecryption,
			Recursive:      recursive,
			MaxResults:     maxResults,
		})
		if err != nil {
			return err
		}

		return client.PrintParameters(params)
	},
}

var ssmPutParameterCmd = &cobra.Command{
	Use:   "put-parameter",
	Short: "Create or update a parameter",
	Long: `Creates or updates a parameter.

Examples:
  omni aws ssm put-parameter --name /app/config --value "config-value" --type String
  omni aws ssm put-parameter --name /app/secret --value "secret" --type SecureString
  omni aws ssm put-parameter --name /app/config --value "new-value" --overwrite`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")
		endpointURL, _ := cmd.Flags().GetString("endpoint-url")

		name, _ := cmd.Flags().GetString("name")
		value, _ := cmd.Flags().GetString("value")
		paramType, _ := cmd.Flags().GetString("type")
		description, _ := cmd.Flags().GetString("description")
		overwrite, _ := cmd.Flags().GetBool("overwrite")
		keyId, _ := cmd.Flags().GetString("key-id")

		opts := awscommon.Options{
			Profile:     profile,
			Region:      region,
			Output:      output,
			EndpointURL: endpointURL,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := ssm.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output), awscommon.GetEndpointURL(opts))

		result, err := client.PutParameter(ctx, ssm.PutParameterInput{
			Name:        name,
			Value:       value,
			Type:        paramType,
			Description: description,
			Overwrite:   overwrite,
			KeyId:       keyId,
		})
		if err != nil {
			return err
		}

		return client.PrintPutParameter(result)
	},
}

var ssmDeleteParameterCmd = &cobra.Command{
	Use:   "delete-parameter",
	Short: "Delete a parameter",
	Long: `Deletes a parameter.

Examples:
  omni aws ssm delete-parameter --name /app/config`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		endpointURL, _ := cmd.Flags().GetString("endpoint-url")

		name, _ := cmd.Flags().GetString("name")

		opts := awscommon.Options{
			Profile:     profile,
			Region:      region,
			EndpointURL: endpointURL,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := ssm.NewClient(cfg, cmd.OutOrStdout(), awscommon.OutputJSON, awscommon.GetEndpointURL(opts))

		if err := client.DeleteParameter(ctx, name); err != nil {
			return err
		}

		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "{}")
		return nil
	},
}

func init() {
	awsCmd.AddCommand(ssmCmd)
	ssmCmd.AddCommand(ssmGetParameterCmd)
	ssmCmd.AddCommand(ssmGetParametersCmd)
	ssmCmd.AddCommand(ssmGetParametersByPathCmd)
	ssmCmd.AddCommand(ssmPutParameterCmd)
	ssmCmd.AddCommand(ssmDeleteParameterCmd)

	// get-parameter flags
	ssmGetParameterCmd.Flags().String("name", "", "parameter name (required)")
	ssmGetParameterCmd.Flags().Bool("with-decryption", false, "decrypt SecureString values")
	_ = ssmGetParameterCmd.MarkFlagRequired("name")

	// get-parameters flags
	ssmGetParametersCmd.Flags().StringSlice("names", nil, "parameter names (required)")
	ssmGetParametersCmd.Flags().Bool("with-decryption", false, "decrypt SecureString values")
	_ = ssmGetParametersCmd.MarkFlagRequired("names")

	// get-parameters-by-path flags
	ssmGetParametersByPathCmd.Flags().String("path", "", "parameter path (required)")
	ssmGetParametersByPathCmd.Flags().Bool("with-decryption", false, "decrypt SecureString values")
	ssmGetParametersByPathCmd.Flags().Bool("recursive", false, "include nested parameters")
	ssmGetParametersByPathCmd.Flags().Int32("max-results", 10, "maximum results per page")
	_ = ssmGetParametersByPathCmd.MarkFlagRequired("path")

	// put-parameter flags
	ssmPutParameterCmd.Flags().String("name", "", "parameter name (required)")
	ssmPutParameterCmd.Flags().String("value", "", "parameter value (required)")
	ssmPutParameterCmd.Flags().String("type", "String", "parameter type: String, StringList, SecureString")
	ssmPutParameterCmd.Flags().String("description", "", "parameter description")
	ssmPutParameterCmd.Flags().Bool("overwrite", false, "overwrite existing parameter")
	ssmPutParameterCmd.Flags().String("key-id", "", "KMS key for SecureString")
	_ = ssmPutParameterCmd.MarkFlagRequired("name")
	_ = ssmPutParameterCmd.MarkFlagRequired("value")

	// delete-parameter flags
	ssmDeleteParameterCmd.Flags().String("name", "", "parameter name (required)")
	_ = ssmDeleteParameterCmd.MarkFlagRequired("name")
}

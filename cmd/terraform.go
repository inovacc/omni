package cmd

import (
	"github.com/inovacc/omni/internal/cli/terraform"
	"github.com/spf13/cobra"
)

var terraformCmd = &cobra.Command{
	Use:     "terraform",
	Aliases: []string{"tf"},
	Short:   "Terraform CLI",
	Long: `Terraform CLI wrapper integrated into omni.

Provides access to all Terraform commands. You can use 'omni terraform'
or the shorter alias 'omni tf'.

Examples:
  omni terraform init
  omni tf plan
  omni tf apply -auto-approve
  omni tf destroy`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.Run(args)
	},
}

// Subcommands for better discoverability

var tfInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Terraform working directory",
	Long: `Initialize a new or existing Terraform working directory.

Examples:
  omni tf init
  omni tf init -upgrade
  omni tf init -reconfigure`,
	RunE: func(cmd *cobra.Command, args []string) error {
		upgrade, _ := cmd.Flags().GetBool("upgrade")
		reconfigure, _ := cmd.Flags().GetBool("reconfigure")
		return terraform.Init(upgrade, reconfigure)
	},
}

var tfPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Create execution plan",
	Long: `Create an execution plan showing what Terraform will do.

Examples:
  omni tf plan
  omni tf plan -out=plan.tfplan
  omni tf plan -var "region=us-east-1"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out, _ := cmd.Flags().GetString("out")
		destroy, _ := cmd.Flags().GetBool("destroy")
		vars, _ := cmd.Flags().GetStringToString("var")
		varFiles, _ := cmd.Flags().GetStringSlice("var-file")
		return terraform.Plan(out, vars, varFiles, destroy)
	},
}

var tfApplyCmd = &cobra.Command{
	Use:   "apply [plan-file]",
	Short: "Apply changes to infrastructure",
	Long: `Apply the changes required to reach the desired state.

Examples:
  omni tf apply
  omni tf apply plan.tfplan
  omni tf apply -auto-approve`,
	RunE: func(cmd *cobra.Command, args []string) error {
		autoApprove, _ := cmd.Flags().GetBool("auto-approve")
		vars, _ := cmd.Flags().GetStringToString("var")
		varFiles, _ := cmd.Flags().GetStringSlice("var-file")
		planFile := ""
		if len(args) > 0 {
			planFile = args[0]
		}
		return terraform.Apply(planFile, autoApprove, vars, varFiles)
	},
}

var tfDestroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy managed infrastructure",
	Long: `Destroy all remote objects managed by Terraform.

Examples:
  omni tf destroy
  omni tf destroy -auto-approve`,
	RunE: func(cmd *cobra.Command, args []string) error {
		autoApprove, _ := cmd.Flags().GetBool("auto-approve")
		vars, _ := cmd.Flags().GetStringToString("var")
		varFiles, _ := cmd.Flags().GetStringSlice("var-file")
		return terraform.Destroy(autoApprove, vars, varFiles)
	},
}

var tfValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long: `Validate the configuration files.

Examples:
  omni tf validate`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.Validate()
	},
}

var tfFmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format configuration files",
	Long: `Reformat configuration files to a canonical format.

Examples:
  omni tf fmt
  omni tf fmt -check
  omni tf fmt -recursive`,
	RunE: func(cmd *cobra.Command, args []string) error {
		check, _ := cmd.Flags().GetBool("check")
		recursive, _ := cmd.Flags().GetBool("recursive")
		diff, _ := cmd.Flags().GetBool("diff")
		return terraform.Fmt(check, recursive, diff)
	},
}

var tfOutputCmd = &cobra.Command{
	Use:   "output [name]",
	Short: "Show output values",
	Long: `Read an output variable from the state file.

Examples:
  omni tf output
  omni tf output instance_ip
  omni tf output -json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		json, _ := cmd.Flags().GetBool("json")
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		return terraform.Output(name, json)
	},
}

// State commands
var tfStateCmd = &cobra.Command{
	Use:   "state",
	Short: "State management commands",
	Long:  `Commands for managing Terraform state.`,
}

var tfStateListCmd = &cobra.Command{
	Use:   "list [addresses...]",
	Short: "List resources in state",
	Long: `List resources in the Terraform state.

Examples:
  omni tf state list
  omni tf state list aws_instance.example`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.StateList(args...)
	},
}

var tfStateShowCmd = &cobra.Command{
	Use:   "show <address>",
	Short: "Show a resource in state",
	Long: `Show the attributes of a single resource in the state.

Examples:
  omni tf state show aws_instance.example`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.StateShow(args[0])
	},
}

var tfStateMvCmd = &cobra.Command{
	Use:   "mv <source> <destination>",
	Short: "Move resource in state",
	Long: `Move a resource from one address to another.

Examples:
  omni tf state mv aws_instance.old aws_instance.new`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		return terraform.StateMv(args[0], args[1], dryRun)
	},
}

var tfStateRmCmd = &cobra.Command{
	Use:   "rm <addresses...>",
	Short: "Remove resources from state",
	Long: `Remove resources from the Terraform state.

Examples:
  omni tf state rm aws_instance.example`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		return terraform.StateRm(args, dryRun)
	},
}

// Workspace commands
var tfWorkspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Workspace management commands",
	Long:  `Commands for managing Terraform workspaces.`,
}

var tfWorkspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workspaces",
	Long: `List all available workspaces.

Examples:
  omni tf workspace list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.WorkspaceList()
	},
}

var tfWorkspaceNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create new workspace",
	Long: `Create a new workspace.

Examples:
  omni tf workspace new development`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.WorkspaceNew(args[0])
	},
}

var tfWorkspaceSelectCmd = &cobra.Command{
	Use:   "select <name>",
	Short: "Select workspace",
	Long: `Select a workspace to use.

Examples:
  omni tf workspace select production`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.WorkspaceSelect(args[0])
	},
}

var tfWorkspaceDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete workspace",
	Long: `Delete a workspace.

Examples:
  omni tf workspace delete old-workspace`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		return terraform.WorkspaceDelete(args[0], force)
	},
}

var tfWorkspaceShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current workspace",
	Long: `Show the name of the current workspace.

Examples:
  omni tf workspace show`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.WorkspaceShow()
	},
}

var tfShowCmd = &cobra.Command{
	Use:   "show [plan-file]",
	Short: "Show plan or state",
	Long: `Show a human-readable output from a plan file or state.

Examples:
  omni tf show
  omni tf show plan.tfplan
  omni tf show -json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		json, _ := cmd.Flags().GetBool("json")
		path := ""
		if len(args) > 0 {
			path = args[0]
		}
		return terraform.Show(path, json)
	},
}

var tfVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show Terraform version",
	Long: `Show the current Terraform version.

Examples:
  omni tf version`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.Version()
	},
}

var tfImportCmd = &cobra.Command{
	Use:   "import <address> <id>",
	Short: "Import existing infrastructure",
	Long: `Import existing infrastructure into Terraform state.

Examples:
  omni tf import aws_instance.example i-1234567890`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vars, _ := cmd.Flags().GetStringToString("var")
		varFiles, _ := cmd.Flags().GetStringSlice("var-file")
		return terraform.Import(args[0], args[1], vars, varFiles)
	},
}

var tfTaintCmd = &cobra.Command{
	Use:   "taint <address>",
	Short: "Mark resource for recreation",
	Long: `Mark a resource instance as not fully functional.

Examples:
  omni tf taint aws_instance.example`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.Taint(args[0])
	},
}

var tfUntaintCmd = &cobra.Command{
	Use:   "untaint <address>",
	Short: "Remove taint from resource",
	Long: `Remove the taint state from a resource instance.

Examples:
  omni tf untaint aws_instance.example`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.Untaint(args[0])
	},
}

var tfRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh state",
	Long: `Update the state file with real-world infrastructure.

Examples:
  omni tf refresh`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vars, _ := cmd.Flags().GetStringToString("var")
		varFiles, _ := cmd.Flags().GetStringSlice("var-file")
		return terraform.Refresh(vars, varFiles)
	},
}

var tfGraphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate dependency graph",
	Long: `Generate a visual representation of dependencies.

Examples:
  omni tf graph
  omni tf graph | dot -Tpng > graph.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		plan, _ := cmd.Flags().GetString("plan")
		drawCycles, _ := cmd.Flags().GetBool("draw-cycles")
		return terraform.Graph(plan, drawCycles)
	},
}

var tfConsoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Interactive console",
	Long: `Launch an interactive console for evaluating expressions.

Examples:
  omni tf console`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.Console()
	},
}

var tfProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Show provider information",
	Long: `Show the providers required for this configuration.

Examples:
  omni tf providers`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.Providers()
	},
}

var tfGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Download modules",
	Long: `Download and install modules for the configuration.

Examples:
  omni tf get
  omni tf get -update`,
	RunE: func(cmd *cobra.Command, args []string) error {
		update, _ := cmd.Flags().GetBool("update")
		return terraform.Get(update)
	},
}

var tfTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests",
	Long: `Execute Terraform test files.

Examples:
  omni tf test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return terraform.Test()
	},
}

func init() {
	rootCmd.AddCommand(terraformCmd)

	// init
	tfInitCmd.Flags().Bool("upgrade", false, "Upgrade modules and plugins")
	tfInitCmd.Flags().Bool("reconfigure", false, "Reconfigure backend")
	terraformCmd.AddCommand(tfInitCmd)

	// plan
	tfPlanCmd.Flags().StringP("out", "o", "", "Write plan to file")
	tfPlanCmd.Flags().Bool("destroy", false, "Create destroy plan")
	tfPlanCmd.Flags().StringToString("var", nil, "Set variables")
	tfPlanCmd.Flags().StringSlice("var-file", nil, "Variable files")
	terraformCmd.AddCommand(tfPlanCmd)

	// apply
	tfApplyCmd.Flags().Bool("auto-approve", false, "Skip interactive approval")
	tfApplyCmd.Flags().StringToString("var", nil, "Set variables")
	tfApplyCmd.Flags().StringSlice("var-file", nil, "Variable files")
	terraformCmd.AddCommand(tfApplyCmd)

	// destroy
	tfDestroyCmd.Flags().Bool("auto-approve", false, "Skip interactive approval")
	tfDestroyCmd.Flags().StringToString("var", nil, "Set variables")
	tfDestroyCmd.Flags().StringSlice("var-file", nil, "Variable files")
	terraformCmd.AddCommand(tfDestroyCmd)

	// validate
	terraformCmd.AddCommand(tfValidateCmd)

	// fmt
	tfFmtCmd.Flags().Bool("check", false, "Check if formatted")
	tfFmtCmd.Flags().Bool("recursive", false, "Process subdirectories")
	tfFmtCmd.Flags().Bool("diff", false, "Display diffs")
	terraformCmd.AddCommand(tfFmtCmd)

	// output
	tfOutputCmd.Flags().Bool("json", false, "JSON output")
	terraformCmd.AddCommand(tfOutputCmd)

	// state
	tfStateCmd.AddCommand(tfStateListCmd)
	tfStateCmd.AddCommand(tfStateShowCmd)
	tfStateMvCmd.Flags().Bool("dry-run", false, "Dry run")
	tfStateCmd.AddCommand(tfStateMvCmd)
	tfStateRmCmd.Flags().Bool("dry-run", false, "Dry run")
	tfStateCmd.AddCommand(tfStateRmCmd)
	terraformCmd.AddCommand(tfStateCmd)

	// workspace
	tfWorkspaceCmd.AddCommand(tfWorkspaceListCmd)
	tfWorkspaceCmd.AddCommand(tfWorkspaceNewCmd)
	tfWorkspaceCmd.AddCommand(tfWorkspaceSelectCmd)
	tfWorkspaceDeleteCmd.Flags().Bool("force", false, "Force delete")
	tfWorkspaceCmd.AddCommand(tfWorkspaceDeleteCmd)
	tfWorkspaceCmd.AddCommand(tfWorkspaceShowCmd)
	terraformCmd.AddCommand(tfWorkspaceCmd)

	// show
	tfShowCmd.Flags().Bool("json", false, "JSON output")
	terraformCmd.AddCommand(tfShowCmd)

	// version
	terraformCmd.AddCommand(tfVersionCmd)

	// import
	tfImportCmd.Flags().StringToString("var", nil, "Set variables")
	tfImportCmd.Flags().StringSlice("var-file", nil, "Variable files")
	terraformCmd.AddCommand(tfImportCmd)

	// taint/untaint
	terraformCmd.AddCommand(tfTaintCmd)
	terraformCmd.AddCommand(tfUntaintCmd)

	// refresh
	tfRefreshCmd.Flags().StringToString("var", nil, "Set variables")
	tfRefreshCmd.Flags().StringSlice("var-file", nil, "Variable files")
	terraformCmd.AddCommand(tfRefreshCmd)

	// graph
	tfGraphCmd.Flags().String("plan", "", "Plan file")
	tfGraphCmd.Flags().Bool("draw-cycles", false, "Draw cycles")
	terraformCmd.AddCommand(tfGraphCmd)

	// console
	terraformCmd.AddCommand(tfConsoleCmd)

	// providers
	terraformCmd.AddCommand(tfProvidersCmd)

	// get
	tfGetCmd.Flags().Bool("update", false, "Update modules")
	terraformCmd.AddCommand(tfGetCmd)

	// test
	terraformCmd.AddCommand(tfTestCmd)
}

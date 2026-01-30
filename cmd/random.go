package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/random"
	"github.com/spf13/cobra"
)

// randomCmd represents the random command
var randomCmd = &cobra.Command{
	Use:   "random [OPTION]...",
	Short: "Generate random values",
	Long: `Generate random numbers, strings, or bytes using crypto/rand.

Types:
  int, integer    random integer (use --min, --max)
  float, decimal  random float between 0 and 1
  string, str     random alphanumeric string (default)
  alpha           random letters only
  alnum           random alphanumeric
  hex             random hexadecimal string
  password        random password (letters, digits, symbols)
  bytes           random bytes as hex
  custom          use custom charset (-c)

  -n, --count N     number of values to generate
  -l, --length N    length of strings (default 16)
  -t, --type TYPE   value type (default: string)
  --min N           minimum for integers
  --max N           maximum for integers
  -c, --charset STR custom character set
  -s, --separator   separator between values (default: newline)

Examples:
  omni random                         # random 16-char string
  omni random -t int --max 100        # random int 0-99
  omni random -t hex -l 32            # random 32-char hex
  omni random -t password -l 20       # random password
  omni random -n 5 -t int --max 10    # 5 random ints 0-9
  omni random -t custom -c "abc123"   # from custom charset`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := random.RandomOptions{}

		opts.Count, _ = cmd.Flags().GetInt("count")
		opts.Length, _ = cmd.Flags().GetInt("length")
		opts.Type, _ = cmd.Flags().GetString("type")
		opts.Min, _ = cmd.Flags().GetInt64("min")
		opts.Max, _ = cmd.Flags().GetInt64("max")
		opts.Charset, _ = cmd.Flags().GetString("charset")
		opts.Sep, _ = cmd.Flags().GetString("separator")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return random.RunRandom(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(randomCmd)

	randomCmd.Flags().IntP("count", "n", 1, "number of values to generate")
	randomCmd.Flags().IntP("length", "l", 16, "length of random strings")
	randomCmd.Flags().StringP("type", "t", "string", "type: int, float, string, alpha, hex, password, bytes, custom")
	randomCmd.Flags().Int64("min", 0, "minimum value for integers")
	randomCmd.Flags().Int64("max", 100, "maximum value for integers")
	randomCmd.Flags().StringP("charset", "c", "", "custom character set")
	randomCmd.Flags().StringP("separator", "s", "\n", "separator between values")
	randomCmd.Flags().Bool("json", false, "output as JSON")
}

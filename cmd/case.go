package cmd

import (
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/caseconv"
	"github.com/inovacc/omni/internal/cli/output"
	"github.com/spf13/cobra"
)

// caseCmd represents the case command
var caseCmd = &cobra.Command{
	Use:   "case",
	Short: "Text case conversion utilities",
	Long: `Convert text between different case conventions.

Subcommands:
  upper     UPPERCASE
  lower     lowercase
  title     Title Case
  sentence  Sentence case
  camel     camelCase
  pascal    PascalCase
  snake     snake_case
  kebab     kebab-case
  constant  CONSTANT_CASE
  dot       dot.case
  path      path/case
  swap      sWAP cASE
  toggle    Toggle first char
  detect    Detect case type
  all       Show all conversions

Examples:
  omni case upper "hello world"       # HELLO WORLD
  omni case camel "hello world"       # helloWorld
  omni case snake "helloWorld"        # hello_world
  echo "hello" | omni case upper      # HELLO`,
}

// caseUpperCmd converts to UPPERCASE
var caseUpperCmd = &cobra.Command{
	Use:     "upper [TEXT...]",
	Aliases: []string{"uppercase", "up"},
	Short:   "Convert to UPPERCASE",
	Long: `Convert text to UPPERCASE.

Examples:
  omni case upper "hello world"       # HELLO WORLD
  omni case upper hello world         # HELLO WORLD
  echo "hello" | omni case upper      # HELLO`,
	RunE: runCaseCmd(caseconv.CaseUpper),
}

// caseLowerCmd converts to lowercase
var caseLowerCmd = &cobra.Command{
	Use:     "lower [TEXT...]",
	Aliases: []string{"lowercase", "low"},
	Short:   "Convert to lowercase",
	Long: `Convert text to lowercase.

Examples:
  omni case lower "HELLO WORLD"       # hello world
  echo "HELLO" | omni case lower      # hello`,
	RunE: runCaseCmd(caseconv.CaseLower),
}

// caseTitleCmd converts to Title Case
var caseTitleCmd = &cobra.Command{
	Use:     "title [TEXT...]",
	Aliases: []string{"titlecase"},
	Short:   "Convert to Title Case",
	Long: `Convert text to Title Case (capitalize first letter of each word).

Examples:
  omni case title "hello world"       # Hello World
  echo "hello world" | omni case title`,
	RunE: runCaseCmd(caseconv.CaseTitle),
}

// caseSentenceCmd converts to Sentence case
var caseSentenceCmd = &cobra.Command{
	Use:     "sentence [TEXT...]",
	Aliases: []string{"sentencecase"},
	Short:   "Convert to Sentence case",
	Long: `Convert text to Sentence case (capitalize first letter only).

Examples:
  omni case sentence "hello world"    # Hello world
  echo "HELLO WORLD" | omni case sentence`,
	RunE: runCaseCmd(caseconv.CaseSentence),
}

// caseCamelCmd converts to camelCase
var caseCamelCmd = &cobra.Command{
	Use:     "camel [TEXT...]",
	Aliases: []string{"camelcase"},
	Short:   "Convert to camelCase",
	Long: `Convert text to camelCase.

Examples:
  omni case camel "hello world"       # helloWorld
  omni case camel "Hello_World"       # helloWorld
  omni case camel "hello-world"       # helloWorld`,
	RunE: runCaseCmd(caseconv.CaseCamel),
}

// casePascalCmd converts to PascalCase
var casePascalCmd = &cobra.Command{
	Use:     "pascal [TEXT...]",
	Aliases: []string{"pascalcase"},
	Short:   "Convert to PascalCase",
	Long: `Convert text to PascalCase.

Examples:
  omni case pascal "hello world"      # HelloWorld
  omni case pascal "hello_world"      # HelloWorld
  omni case pascal "hello-world"      # HelloWorld`,
	RunE: runCaseCmd(caseconv.CasePascal),
}

// caseSnakeCmd converts to snake_case
var caseSnakeCmd = &cobra.Command{
	Use:     "snake [TEXT...]",
	Aliases: []string{"snakecase", "snake_case"},
	Short:   "Convert to snake_case",
	Long: `Convert text to snake_case.

Examples:
  omni case snake "hello world"       # hello_world
  omni case snake "helloWorld"        # hello_world
  omni case snake "HelloWorld"        # hello_world`,
	RunE: runCaseCmd(caseconv.CaseSnake),
}

// caseKebabCmd converts to kebab-case
var caseKebabCmd = &cobra.Command{
	Use:     "kebab [TEXT...]",
	Aliases: []string{"kebabcase", "kebab-case"},
	Short:   "Convert to kebab-case",
	Long: `Convert text to kebab-case.

Examples:
  omni case kebab "hello world"       # hello-world
  omni case kebab "helloWorld"        # hello-world
  omni case kebab "HelloWorld"        # hello-world`,
	RunE: runCaseCmd(caseconv.CaseKebab),
}

// caseConstantCmd converts to CONSTANT_CASE
var caseConstantCmd = &cobra.Command{
	Use:     "constant [TEXT...]",
	Aliases: []string{"constantcase", "screaming", "screaming_snake"},
	Short:   "Convert to CONSTANT_CASE",
	Long: `Convert text to CONSTANT_CASE (SCREAMING_SNAKE_CASE).

Examples:
  omni case constant "hello world"    # HELLO_WORLD
  omni case constant "helloWorld"     # HELLO_WORLD`,
	RunE: runCaseCmd(caseconv.CaseConstant),
}

// caseDotCmd converts to dot.case
var caseDotCmd = &cobra.Command{
	Use:     "dot [TEXT...]",
	Aliases: []string{"dotcase"},
	Short:   "Convert to dot.case",
	Long: `Convert text to dot.case.

Examples:
  omni case dot "hello world"         # hello.world
  omni case dot "helloWorld"          # hello.world`,
	RunE: runCaseCmd(caseconv.CaseDot),
}

// casePathCmd converts to path/case
var casePathCmd = &cobra.Command{
	Use:     "path [TEXT...]",
	Aliases: []string{"pathcase"},
	Short:   "Convert to path/case",
	Long: `Convert text to path/case.

Examples:
  omni case path "hello world"        # hello/world
  omni case path "helloWorld"         # hello/world`,
	RunE: runCaseCmd(caseconv.CasePath),
}

// caseSwapCmd swaps case
var caseSwapCmd = &cobra.Command{
	Use:     "swap [TEXT...]",
	Aliases: []string{"swapcase"},
	Short:   "Swap case of each character",
	Long: `Swap the case of each character (upper becomes lower, lower becomes upper).

Examples:
  omni case swap "Hello World"        # hELLO wORLD
  omni case swap "helloWorld"         # HELLOwORLD`,
	RunE: runCaseCmd(caseconv.CaseSwap),
}

// caseToggleCmd toggles first character
var caseToggleCmd = &cobra.Command{
	Use:   "toggle [TEXT...]",
	Short: "Toggle first character's case",
	Long: `Toggle the case of the first character.

Examples:
  omni case toggle "hello"            # Hello
  omni case toggle "Hello"            # hello`,
	RunE: runCaseCmd(caseconv.CaseToggle),
}

// caseDetectCmd detects the case type
var caseDetectCmd = &cobra.Command{
	Use:   "detect [TEXT...]",
	Short: "Detect the case type of text",
	Long: `Detect the case type of the input text.

Examples:
  omni case detect "helloWorld"       # camel
  omni case detect "hello_world"      # snake
  omni case detect "HELLO_WORLD"      # constant`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format := getOutputOpts(cmd).GetFormat()
		f := output.New(cmd.OutOrStdout(), format)

		if len(args) == 0 {
			return runCaseDetectStdin(cmd.OutOrStdout(), f)
		}

		for _, arg := range args {
			ct := caseconv.DetectCase(arg)
			if f.IsJSON() {
				return f.Print(struct {
					Input string `json:"input"`
					Case  string `json:"case"`
				}{Input: arg, Case: string(ct)})
			}
			_, _ = cmd.OutOrStdout().Write([]byte(string(ct) + "\n"))
		}
		return nil
	},
}

// caseAllCmd shows all case conversions
var caseAllCmd = &cobra.Command{
	Use:   "all [TEXT...]",
	Short: "Show all case conversions",
	Long: `Convert text to all supported case types and display results.

Examples:
  omni case all "hello world"
  omni case all "helloWorld"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format := getOutputOpts(cmd).GetFormat()
		f := output.New(cmd.OutOrStdout(), format)
		w := cmd.OutOrStdout()

		if len(args) == 0 {
			return runCaseAllStdin(w, f)
		}

		for _, arg := range args {
			conversions := caseconv.ConvertAll(arg)

			if f.IsJSON() {
				result := struct {
					Input       string            `json:"input"`
					Conversions map[string]string `json:"conversions"`
				}{
					Input:       arg,
					Conversions: make(map[string]string),
				}
				for ct, val := range conversions {
					result.Conversions[string(ct)] = val
				}
				return f.Print(result)
			}

			_, _ = w.Write([]byte("Input: " + arg + "\n"))
			_, _ = w.Write([]byte("───────────────────────────\n"))
			for _, ct := range caseconv.ValidCaseTypes() {
				_, _ = w.Write([]byte(padRight(string(ct), 12) + conversions[ct] + "\n"))
			}
			_, _ = w.Write([]byte("\n"))
		}
		return nil
	},
}

func runCaseCmd(caseType caseconv.CaseType) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		opts := caseconv.Options{
			Case:         caseType,
			OutputFormat: getOutputOpts(cmd).GetFormat(),
		}

		return caseconv.RunCase(cmd.OutOrStdout(), args, opts)
	}
}

func runCaseDetectStdin(w io.Writer, f *output.Formatter) error {
	buf := make([]byte, 4096)
	n, _ := os.Stdin.Read(buf)
	input := strings.TrimSpace(string(buf[:n]))

	ct := caseconv.DetectCase(input)
	if f.IsJSON() {
		return f.Print(struct {
			Input string `json:"input"`
			Case  string `json:"case"`
		}{Input: input, Case: string(ct)})
	}

	_, _ = w.Write([]byte(string(ct) + "\n"))
	return nil
}

func runCaseAllStdin(w io.Writer, f *output.Formatter) error {
	buf := make([]byte, 4096)
	n, _ := os.Stdin.Read(buf)
	input := strings.TrimSpace(string(buf[:n]))

	conversions := caseconv.ConvertAll(input)

	if f.IsJSON() {
		result := struct {
			Input       string            `json:"input"`
			Conversions map[string]string `json:"conversions"`
		}{
			Input:       input,
			Conversions: make(map[string]string),
		}
		for ct, val := range conversions {
			result.Conversions[string(ct)] = val
		}
		return f.Print(result)
	}

	_, _ = w.Write([]byte("Input: " + input + "\n"))
	_, _ = w.Write([]byte("───────────────────────────\n"))
	for _, ct := range caseconv.ValidCaseTypes() {
		_, _ = w.Write([]byte(padRight(string(ct), 12) + conversions[ct] + "\n"))
	}
	return nil
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}

	return s + strings.Repeat(" ", length-len(s))
}

func init() {
	rootCmd.AddCommand(caseCmd)

	// Add subcommands
	caseCmd.AddCommand(caseUpperCmd)
	caseCmd.AddCommand(caseLowerCmd)
	caseCmd.AddCommand(caseTitleCmd)
	caseCmd.AddCommand(caseSentenceCmd)
	caseCmd.AddCommand(caseCamelCmd)
	caseCmd.AddCommand(casePascalCmd)
	caseCmd.AddCommand(caseSnakeCmd)
	caseCmd.AddCommand(caseKebabCmd)
	caseCmd.AddCommand(caseConstantCmd)
	caseCmd.AddCommand(caseDotCmd)
	caseCmd.AddCommand(casePathCmd)
	caseCmd.AddCommand(caseSwapCmd)
	caseCmd.AddCommand(caseToggleCmd)
	caseCmd.AddCommand(caseDetectCmd)
	caseCmd.AddCommand(caseAllCmd)

}

//nolint:forbidigo,nilerr // Borrowed code from twig
package confirm

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/inovacc/omni/internal/twig/models"
	"github.com/manifoldco/promptui"
)

// ShowPreview displays a formatted preview of the directory structure that will be created,
// including statistics and warnings. It prompts the user for confirmation unless dryRun is true.
// Returns true if the user confirms, false otherwise.
func ShowPreview(root *models.Node, targetPath string, dryRun bool) (bool, error) {
	stats := models.CalculateStats(root)

	// Print header
	printHeader("Structure Creation Preview")

	// Print target info
	fmt.Printf("\n")
	color.Cyan("Target directory: %s", targetPath)
	fmt.Printf("Source: stdin/file\n")

	// Print summary
	fmt.Printf("\n")
	printSummary(stats)

	// Print preview of structure (first 15 items)
	fmt.Printf("\n")
	printTreePreview(root, 15)

	// Print warnings
	printWarnings(targetPath, stats)

	// Ask for confirmation if not dry run
	if dryRun {
		fmt.Println("\n✓ Dry run mode - no changes will be made")
		return true, nil
	}

	return askConfirmation()
}

func printHeader(title string) {
	width := 60
	border := strings.Repeat("═", width-2)

	fmt.Printf("╔%s╗\n", border)

	padding := (width - len(title) - 2) / 2
	fmt.Printf("║%s%s%s║\n",
		strings.Repeat(" ", padding),
		title,
		strings.Repeat(" ", width-padding-len(title)-2))
	fmt.Printf("╚%s╝\n", border)
}

func printSummary(stats *models.TreeStats) {
	fmt.Println("Summary:")
	fmt.Printf("  • Directories to create: %s\n", color.GreenString("%d", stats.TotalDirs))
	fmt.Printf("  • Files to create: %s\n", color.GreenString("%d", stats.TotalFiles))
	fmt.Printf("  • Total operations: %s\n", color.CyanString("%d", stats.TotalDirs+stats.TotalFiles))
	fmt.Printf("  • Max depth: %s\n", color.YellowString("%d", stats.MaxDepth))
}

func printTreePreview(root *models.Node, maxItems int) {
	fmt.Println("Preview (first items):")

	count := 0
	printTreeRecursive(root, "", true, &count, maxItems)

	stats := models.CalculateStats(root)

	total := stats.TotalDirs + stats.TotalFiles
	if total > maxItems {
		fmt.Printf("  ... and %d more items\n", total-maxItems)
	}
}

func printTreeRecursive(node *models.Node, prefix string, isLast bool, count *int, maxItems int) {
	if *count >= maxItems {
		return
	}

	if node == nil {
		return
	}

	// Skip root in preview
	if prefix == "" {
		for _, child := range node.Children {
			printTreeRecursive(child, "", false, count, maxItems)
		}

		return
	}

	// Print current node
	marker := "├──"
	if isLast {
		marker = "└──"
	}

	icon := "✓"

	name := node.Name
	if node.IsDir {
		name = color.BlueString(name + "/")
	}

	fmt.Printf("  %s %s %s\n", marker, color.GreenString(icon), name)

	*count++

	if *count >= maxItems {
		return
	}

	// Print children
	for i, child := range node.Children {
		isChildLast := i == len(node.Children)-1

		childPrefix := prefix
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}

		printTreeRecursive(child, childPrefix, isChildLast, count, maxItems)

		if *count >= maxItems {
			return
		}
	}
}

func printWarnings(targetPath string, stats *models.TreeStats) {
	fmt.Printf("\n")
	color.Yellow("⚠ Warning: Directory '%s' will be modified", targetPath)

	if stats.TotalDirs+stats.TotalFiles > 100 {
		color.Yellow("⚠ Warning: Large structure (%d items) will be created", stats.TotalDirs+stats.TotalFiles)
	}
}

func askConfirmation() (bool, error) {
	prompt := promptui.Prompt{
		Label:     "Continue with creation",
		IsConfirm: true,
		Default:   "N",
	}

	result, err := prompt.Run()
	if err != nil {
		// User pressed Ctrl+C or answered No
		return false, nil
	}

	// Check if user confirmed
	result = strings.ToLower(strings.TrimSpace(result))

	return result == "y" || result == "yes", nil
}

// AskForce prompts the user for confirmation when the --force flag is used.
// It displays a prominent warning about force mode being enabled and asks for
// explicit confirmation before proceeding. Returns true if confirmed, false otherwise.
func AskForce(targetPath string) (bool, error) {
	color.Red("\n⚠ WARNING: Force mode is enabled!")
	fmt.Printf("This will create/overwrite files at: %s\n", targetPath)

	prompt := promptui.Prompt{
		Label:     "Are you absolutely sure",
		IsConfirm: true,
		Default:   "N",
	}

	result, err := prompt.Run()
	if err != nil {
		return false, nil
	}

	result = strings.ToLower(strings.TrimSpace(result))

	return result == "y" || result == "yes", nil
}

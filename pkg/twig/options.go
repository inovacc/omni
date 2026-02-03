package twig

import (
	"github.com/inovacc/omni/pkg/twig/builder"
	"github.com/inovacc/omni/pkg/twig/formatter"
	"github.com/inovacc/omni/pkg/twig/scanner"
)

// TreeOption is a function that configures a Tree instance.
type TreeOption func(*Tree)

// Scanner Options

// WithMaxDepth sets the maximum depth to scan (-1 for unlimited).
func WithMaxDepth(depth int) TreeOption {
	return func(t *Tree) {
		t.scanConfig.MaxDepth = depth
	}
}

// WithShowHidden enables or disables showing hidden files.
func WithShowHidden(show bool) TreeOption {
	return func(t *Tree) {
		t.scanConfig.ShowHidden = show
	}
}

// WithIgnorePatterns sets patterns to ignore during scanning.
// Patterns use filepath.Match syntax (e.g., "*.log", "node_modules").
func WithIgnorePatterns(patterns ...string) TreeOption {
	return func(t *Tree) {
		t.scanConfig.IgnorePatterns = append(t.scanConfig.IgnorePatterns, patterns...)
	}
}

// WithDirsOnly sets whether to show only directories (no files).
func WithDirsOnly(dirsOnly bool) TreeOption {
	return func(t *Tree) {
		t.scanConfig.DirsOnly = dirsOnly
	}
}

// WithScanConfig sets a custom scan configuration.
func WithScanConfig(config *scanner.ScanConfig) TreeOption {
	return func(t *Tree) {
		t.scanConfig = config
	}
}

// Formatter Options

// WithColors enables or disables colored output.
func WithColors(show bool) TreeOption {
	return func(t *Tree) {
		t.formatConfig.ShowColors = show
	}
}

// WithDirSlash enables or disables trailing slash on directory names.
func WithDirSlash(show bool) TreeOption {
	return func(t *Tree) {
		t.formatConfig.ShowDirSlash = show
	}
}

// WithShowSize enables or disables showing file sizes.
func WithShowSize(show bool) TreeOption {
	return func(t *Tree) {
		t.formatConfig.ShowSize = show
	}
}

// WithShowDate enables or disables showing modification dates.
func WithShowDate(show bool) TreeOption {
	return func(t *Tree) {
		t.formatConfig.ShowDate = show
	}
}

// WithShowHash enables or disables showing file hashes.
func WithShowHash(show bool) TreeOption {
	return func(t *Tree) {
		t.scanConfig.ShowHash = show
		t.formatConfig.ShowHash = show
	}
}

// WithFlattenFilesHash enables or disables flattened files hash output.
func WithFlattenFilesHash(flatten bool) TreeOption {
	return func(t *Tree) {
		t.scanConfig.ShowHash = flatten // Auto-enable hash calculation
		t.formatConfig.FlattenFilesHash = flatten
	}
}

// WithJSONOutput enables or disables JSON output format.
func WithJSONOutput(json bool) TreeOption {
	return func(t *Tree) {
		t.formatConfig.JSONOutput = json
	}
}

// WithFormatConfig sets a custom format configuration.
func WithFormatConfig(config *formatter.FormatConfig) TreeOption {
	return func(t *Tree) {
		t.formatConfig = config
	}
}

// Builder Options

// WithDryRun enables or disables dry-run mode (show what would be created without creating).
func WithDryRun(dryRun bool) TreeOption {
	return func(t *Tree) {
		t.buildConfig.DryRun = dryRun
	}
}

// WithOverwrite enables or disables overwriting existing files.
func WithOverwrite(overwrite bool) TreeOption {
	return func(t *Tree) {
		t.buildConfig.Overwrite = overwrite
	}
}

// WithSkipExisting sets whether to skip existing files/directories.
func WithSkipExisting(skip bool) TreeOption {
	return func(t *Tree) {
		t.buildConfig.SkipExisting = skip
	}
}

// WithAbortOnConflict sets whether to abort if target directory already exists.
func WithAbortOnConflict(abort bool) TreeOption {
	return func(t *Tree) {
		t.buildConfig.AbortOnConflict = abort
	}
}

// WithVerbose enables or disables verbose output during building.
func WithVerbose(verbose bool) TreeOption {
	return func(t *Tree) {
		t.buildConfig.Verbose = verbose
	}
}

// WithBuildConfig sets a custom build configuration.
func WithBuildConfig(config *builder.BuildConfig) TreeOption {
	return func(t *Tree) {
		t.buildConfig = config
	}
}

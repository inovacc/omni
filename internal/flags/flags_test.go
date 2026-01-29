package flags

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testFeatureName generates a unique feature name for testing
func testFeatureName(t *testing.T, suffix string) string {
	t.Helper()
	// Use test name to create unique feature names
	name := strings.ReplaceAll(t.Name(), "/", "_")
	name = strings.ReplaceAll(name, " ", "_")

	return "TEST_" + name + "_" + suffix
}

// cleanupFeature removes test feature files
func cleanupFeature(t *testing.T, feature string) {
	t.Helper()

	if err := ensureInit(); err != nil {
		return
	}

	feature = strings.ToUpper(feature)
	enabled := filepath.Join(appDir, prefix+feature+"_ENABLED")
	disabled := filepath.Join(appDir, prefix+feature+"_DISABLED")

	_ = os.Remove(enabled)
	_ = os.Remove(disabled)
}

func TestEnableFeature(t *testing.T) {
	feature := testFeatureName(t, "ENABLE")
	t.Cleanup(func() { cleanupFeature(t, feature) })

	// Ensure init is called first (appDir needs to be set)
	if err := ensureInit(); err != nil {
		t.Fatalf("ensureInit() failed: %v", err)
	}

	// Enable a new feature
	if err := EnableFeature(feature, ""); err != nil {
		t.Fatalf("EnableFeature() failed: %v", err)
	}

	// Verify file exists
	enabledPath := filepath.Join(appDir, prefix+strings.ToUpper(feature)+"_ENABLED")
	if _, err := os.Stat(enabledPath); os.IsNotExist(err) {
		t.Error("expected enabled file to exist")
	}
}

func TestDisableFeature(t *testing.T) {
	feature := testFeatureName(t, "DISABLE")
	t.Cleanup(func() { cleanupFeature(t, feature) })

	// Disable a new feature
	if err := DisableFeature(feature); err != nil {
		t.Fatalf("DisableFeature() failed: %v", err)
	}

	// Verify file exists
	if err := ensureInit(); err != nil {
		t.Fatalf("ensureInit() failed: %v", err)
	}

	disabledPath := filepath.Join(appDir, prefix+strings.ToUpper(feature)+"_DISABLED")
	if _, err := os.Stat(disabledPath); os.IsNotExist(err) {
		t.Error("expected disabled file to exist")
	}
}

func TestEnableDisableToggle(t *testing.T) {
	feature := testFeatureName(t, "TOGGLE")
	t.Cleanup(func() { cleanupFeature(t, feature) })

	if err := ensureInit(); err != nil {
		t.Fatalf("ensureInit() failed: %v", err)
	}

	enabledPath := filepath.Join(appDir, prefix+strings.ToUpper(feature)+"_ENABLED")
	disabledPath := filepath.Join(appDir, prefix+strings.ToUpper(feature)+"_DISABLED")

	// Start by enabling
	if err := EnableFeature(feature, ""); err != nil {
		t.Fatalf("EnableFeature() failed: %v", err)
	}

	if _, err := os.Stat(enabledPath); os.IsNotExist(err) {
		t.Error("expected enabled file after EnableFeature")
	}

	// Now disable (should rename)
	if err := DisableFeature(feature); err != nil {
		t.Fatalf("DisableFeature() failed: %v", err)
	}

	if _, err := os.Stat(disabledPath); os.IsNotExist(err) {
		t.Error("expected disabled file after DisableFeature")
	}

	if _, err := os.Stat(enabledPath); !os.IsNotExist(err) {
		t.Error("enabled file should not exist after DisableFeature")
	}

	// Re-enable (should rename back)
	if err := EnableFeature(feature, ""); err != nil {
		t.Fatalf("EnableFeature() failed on re-enable: %v", err)
	}

	if _, err := os.Stat(enabledPath); os.IsNotExist(err) {
		t.Error("expected enabled file after re-enabling")
	}

	if _, err := os.Stat(disabledPath); !os.IsNotExist(err) {
		t.Error("disabled file should not exist after re-enabling")
	}
}

func TestLoadFeatureFlags(t *testing.T) {
	feature1 := testFeatureName(t, "LOAD1")
	feature2 := testFeatureName(t, "LOAD2")
	t.Cleanup(func() {
		cleanupFeature(t, feature1)
		cleanupFeature(t, feature2)
	})

	// Create one enabled and one disabled feature
	if err := EnableFeature(feature1, ""); err != nil {
		t.Fatalf("EnableFeature() failed: %v", err)
	}

	if err := DisableFeature(feature2); err != nil {
		t.Fatalf("DisableFeature() failed: %v", err)
	}

	// Load flags
	flags, err := LoadFeatureFlags()
	if err != nil {
		t.Fatalf("LoadFeatureFlags() failed: %v", err)
	}

	// Check feature1 is enabled
	f1Key := strings.ToUpper(feature1)
	if enabled, exists := flags[f1Key]; !exists || !enabled {
		t.Errorf("expected %s to be enabled, got exists=%v enabled=%v", f1Key, exists, enabled)
	}

	// Check feature2 is disabled
	f2Key := strings.ToUpper(feature2)
	if enabled, exists := flags[f2Key]; !exists || enabled {
		t.Errorf("expected %s to be disabled, got exists=%v enabled=%v", f2Key, exists, enabled)
	}
}

func TestIsFeatureEnabled(t *testing.T) {
	feature := testFeatureName(t, "ISENABLED")
	t.Cleanup(func() { cleanupFeature(t, feature) })

	// Initially not set
	if IsFeatureEnabled(feature) {
		t.Error("expected feature to not be enabled initially")
	}

	// Enable it
	if err := EnableFeature(feature, ""); err != nil {
		t.Fatalf("EnableFeature() failed: %v", err)
	}

	// Force cache refresh
	_, _ = LoadFeatureFlags()

	if !IsFeatureEnabled(feature) {
		t.Error("expected feature to be enabled after EnableFeature")
	}
}

func TestIsFeatureDisabled(t *testing.T) {
	feature := testFeatureName(t, "ISDISABLED")
	t.Cleanup(func() { cleanupFeature(t, feature) })

	// Initially not set (IsFeatureDisabled returns false for non-existent)
	if IsFeatureDisabled(feature) {
		t.Error("expected feature to not be disabled initially (not set)")
	}

	// Disable it
	if err := DisableFeature(feature); err != nil {
		t.Fatalf("DisableFeature() failed: %v", err)
	}

	// Force cache refresh
	_, _ = LoadFeatureFlags()

	if !IsFeatureDisabled(feature) {
		t.Error("expected feature to be disabled after DisableFeature")
	}
}

func TestIsFeatureSet(t *testing.T) {
	feature := testFeatureName(t, "ISSET")
	t.Cleanup(func() { cleanupFeature(t, feature) })

	// Initially not set
	if IsFeatureSet(feature) {
		t.Error("expected feature to not be set initially")
	}

	// Enable it (now it's set)
	if err := EnableFeature(feature, ""); err != nil {
		t.Fatalf("EnableFeature() failed: %v", err)
	}

	// Force cache refresh
	_, _ = LoadFeatureFlags()

	if !IsFeatureSet(feature) {
		t.Error("expected feature to be set after EnableFeature")
	}
}

func TestGetCachedFlags(t *testing.T) {
	feature := testFeatureName(t, "CACHED")
	t.Cleanup(func() { cleanupFeature(t, feature) })

	// Enable a feature and load to populate cache
	if err := EnableFeature(feature, ""); err != nil {
		t.Fatalf("EnableFeature() failed: %v", err)
	}

	_, _ = LoadFeatureFlags()

	// Get cached flags
	cached := GetCachedFlags()
	if cached == nil {
		t.Fatal("expected cached flags to not be nil")
	}

	// Verify our feature is in cache
	f := strings.ToUpper(feature)
	if enabled, exists := cached[f]; !exists || !enabled {
		t.Errorf("expected %s to be in cache and enabled", f)
	}

	// Verify returned map is a copy (mutation doesn't affect original)
	cached[f] = false
	cached2 := GetCachedFlags()

	if !cached2[f] {
		t.Error("modifying returned map should not affect cache")
	}
}

func TestFeatureNameCaseInsensitive(t *testing.T) {
	feature := testFeatureName(t, "CASE")
	t.Cleanup(func() { cleanupFeature(t, feature) })

	// Enable with lowercase
	if err := EnableFeature(strings.ToLower(feature), ""); err != nil {
		t.Fatalf("EnableFeature() failed: %v", err)
	}

	// Force cache refresh
	_, _ = LoadFeatureFlags()

	// Check with various cases
	if !IsFeatureEnabled(strings.ToLower(feature)) {
		t.Error("expected feature to be enabled (lowercase check)")
	}

	if !IsFeatureEnabled(strings.ToUpper(feature)) {
		t.Error("expected feature to be enabled (uppercase check)")
	}

	if !IsFeatureEnabled(feature) {
		t.Error("expected feature to be enabled (original case check)")
	}
}

func TestEnabledTakesPrecedenceOverDisabled(t *testing.T) {
	feature := testFeatureName(t, "PRECEDENCE")
	t.Cleanup(func() { cleanupFeature(t, feature) })

	if err := ensureInit(); err != nil {
		t.Fatalf("ensureInit() failed: %v", err)
	}

	// Create both enabled and disabled files (edge case)
	featureUpper := strings.ToUpper(feature)
	enabledPath := filepath.Join(appDir, prefix+featureUpper+"_ENABLED")
	disabledPath := filepath.Join(appDir, prefix+featureUpper+"_DISABLED")

	f1, err := os.Create(enabledPath)
	if err != nil {
		t.Fatalf("failed to create enabled file: %v", err)
	}

	_ = f1.Close()

	f2, err := os.Create(disabledPath)
	if err != nil {
		t.Fatalf("failed to create disabled file: %v", err)
	}

	_ = f2.Close()

	// Load flags - enabled should take precedence
	flags, err := LoadFeatureFlags()
	if err != nil {
		t.Fatalf("LoadFeatureFlags() failed: %v", err)
	}

	if !flags[featureUpper] {
		t.Error("expected enabled to take precedence over disabled")
	}
}

func TestIgnoresDirectories(t *testing.T) {
	if err := ensureInit(); err != nil {
		t.Fatalf("ensureInit() failed: %v", err)
	}

	// Create a directory that looks like a feature flag
	dirName := prefix + "TESTDIR_ENABLED"
	dirPath := filepath.Join(appDir, dirName)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	t.Cleanup(func() { _ = os.Remove(dirPath) })

	// Load flags - should not include the directory
	flags, err := LoadFeatureFlags()
	if err != nil {
		t.Fatalf("LoadFeatureFlags() failed: %v", err)
	}

	if _, exists := flags["TESTDIR"]; exists {
		t.Error("directories should be ignored")
	}
}

func TestShouldIgnoreCommand(t *testing.T) {
	// logger is pre-registered as ignored
	if !ShouldIgnoreCommand("logger") {
		t.Error("expected logger to be ignored")
	}

	if !ShouldIgnoreCommand("LOGGER") {
		t.Error("expected LOGGER to be ignored (case insensitive)")
	}

	// Random command should not be ignored
	if ShouldIgnoreCommand("cat") {
		t.Error("cat should not be ignored")
	}

	// Add a new ignored command
	_ = IgnoreCommand("myCustomCmd")

	if !ShouldIgnoreCommand("mycustomcmd") {
		t.Error("expected mycustomcmd to be ignored after IgnoreCommand")
	}
}

func TestIgnoresNonPrefixedFiles(t *testing.T) {
	if err := ensureInit(); err != nil {
		t.Fatalf("ensureInit() failed: %v", err)
	}

	// Create a file without the OMNI_ prefix
	fileName := "RANDOM_FEATURE_ENABLED"
	filePath := filepath.Join(appDir, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_ = f.Close()

	t.Cleanup(func() { _ = os.Remove(filePath) })

	// Load flags - should not include the non-prefixed file
	flags, err := LoadFeatureFlags()
	if err != nil {
		t.Fatalf("LoadFeatureFlags() failed: %v", err)
	}

	if _, exists := flags["RANDOM_FEATURE"]; exists {
		t.Error("non-prefixed files should be ignored")
	}
}

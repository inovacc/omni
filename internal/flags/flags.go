package flags

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	appName = "omni"
	prefix  = "OMNI_"
)

var (
	appDir    string
	initOnce  sync.Once
	initErr   error
	exportMu  sync.Once
	cacheMu   sync.RWMutex
	flagCache map[string]bool
)

func ensureInit() error {
	initOnce.Do(func() {
		dataDir, err := os.UserCacheDir()
		if err != nil {
			initErr = fmt.Errorf("failed to get user cache dir: %w", err)
			return
		}

		appDir = filepath.Join(dataDir, appName)

		if err := os.MkdirAll(appDir, 0755); err != nil {
			initErr = fmt.Errorf("failed to create app dir: %w", err)
		}
	})

	return initErr
}

func LoadFeatureFlags() (map[string]bool, error) {
	if err := ensureInit(); err != nil {
		return nil, err
	}

	flags := make(map[string]bool)

	entries, err := os.ReadDir(appDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read feature flags dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if !strings.HasPrefix(name, prefix) {
			continue
		}

		switch {
		case strings.HasSuffix(name, "_ENABLED"):
			feature := strings.TrimSuffix(strings.TrimPrefix(name, prefix), "_ENABLED")
			flags[feature] = true

		case strings.HasSuffix(name, "_DISABLED"):
			feature := strings.TrimSuffix(strings.TrimPrefix(name, prefix), "_DISABLED")
			if _, exists := flags[feature]; !exists {
				flags[feature] = false
			}
		}
	}

	// Update cache
	cacheMu.Lock()

	flagCache = flags

	cacheMu.Unlock()

	return flags, nil
}

// GetCachedFlags returns cached flags without reading from disk.
// Returns nil if the cache is not populated. Call LoadFeatureFlags first.
func GetCachedFlags() map[string]bool {
	cacheMu.RLock()
	defer cacheMu.RUnlock()

	if flagCache == nil {
		return nil
	}

	// Return a copy to prevent mutation
	result := make(map[string]bool, len(flagCache))
	maps.Copy(result, flagCache)

	return result
}

func ExportFlagsToEnv() error {
	var err error

	exportMu.Do(func() {
		err = exportFlagsToEnv()
	})

	return err
}

func exportFlagsToEnv() error {
	flags, err := LoadFeatureFlags()
	if err != nil {
		return err
	}

	var errs []string

	for feature, enabled := range flags {
		envKey := prefix + strings.ToUpper(feature)

		value := "0"
		if enabled {
			value = "1"
		}

		if err := os.Setenv(envKey, value); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", envKey, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to set env vars: %s", strings.Join(errs, "; "))
	}

	return nil
}

// getFlags returns cached flags or loads from disk if cache is empty.
func getFlags() map[string]bool {
	if cached := GetCachedFlags(); cached != nil {
		return cached
	}

	flags, err := LoadFeatureFlags()
	if err != nil {
		return nil
	}

	return flags
}

func IsFeatureSet(feature string) bool {
	flags := getFlags()
	if flags == nil {
		return false
	}

	_, exists := flags[strings.ToUpper(feature)]

	return exists
}

func IsFeatureDisabled(feature string) bool {
	flags := getFlags()
	if flags == nil {
		return false
	}

	enabled, exists := flags[strings.ToUpper(feature)]

	return exists && !enabled
}

func IsFeatureEnabled(feature string) bool {
	flags := getFlags()
	if flags == nil {
		return false
	}

	return flags[strings.ToUpper(feature)]
}

func EnableFeature(feature string) error {
	feature = strings.ToUpper(feature)

	disabled := filepath.Join(appDir, fmt.Sprintf("%s%s_DISABLED", prefix, feature))
	enabled := filepath.Join(appDir, fmt.Sprintf("%s%s_ENABLED", prefix, feature))

	// si no existe ninguno, creamos ENABLED
	if _, err := os.Stat(disabled); os.IsNotExist(err) {
		f, err := os.Create(enabled)
		if err != nil {
			return err
		}

		return f.Close()
	}

	return os.Rename(disabled, enabled)
}

func DisableFeature(feature string) error {
	feature = strings.ToUpper(feature)

	enabled := filepath.Join(appDir, fmt.Sprintf("%s%s_ENABLED", prefix, feature))
	disabled := filepath.Join(appDir, fmt.Sprintf("%s%s_DISABLED", prefix, feature))

	if _, err := os.Stat(enabled); os.IsNotExist(err) {
		f, err := os.Create(disabled)
		if err != nil {
			return err
		}

		return f.Close()
	}

	return os.Rename(enabled, disabled)
}

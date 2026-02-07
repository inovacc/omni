package extractor

import "sync"

var (
	mu         sync.RWMutex
	extractors []Extractor
)

// Register adds an extractor to the global registry.
// Typically called from an extractor package's init() function.
func Register(e Extractor) {
	mu.Lock()
	defer mu.Unlock()

	extractors = append(extractors, e)
}

// Match returns the first extractor that can handle the given URL.
// The generic extractor (if registered) is always tried last.
func Match(url string) (Extractor, bool) {
	mu.RLock()
	defer mu.RUnlock()

	var generic Extractor

	for _, e := range extractors {
		if e.Name() == "Generic" {
			generic = e
			continue
		}

		if e.Suitable(url) {
			return e, true
		}
	}
	// Fall back to generic.
	if generic != nil && generic.Suitable(url) {
		return generic, true
	}

	return nil, false
}

// All returns all registered extractors.
func All() []Extractor {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]Extractor, len(extractors))
	copy(result, extractors)

	return result
}

// Names returns the names of all registered extractors.
func Names() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, len(extractors))
	for i, e := range extractors {
		names[i] = e.Name()
	}

	return names
}

// Package extractor defines the Extractor interface and provides a global
// registry for site-specific video metadata extractors. Extractors register
// themselves via init() and are matched against URLs at runtime.
package extractor

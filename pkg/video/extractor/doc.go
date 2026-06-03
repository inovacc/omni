// Package extractor defines the Extractor interface and provides a global
// registry for site-specific video metadata extractors. Extractors register
// themselves via init() and are matched against URLs at runtime.
//
// Experimental: this package's API may change before a stable release and is
// not covered by the v1.0 compatibility guarantee. It tracks third-party site
// internals (YouTube/innertube/HLS) and will change as those change.
package extractor

// Package format provides video format sorting by quality and a selector
// that parses format specification strings like "best", "worst",
// "bestvideo", and filter expressions like "best[height<=720]".
//
// Experimental: this package's API may change before a stable release and is
// not covered by the v1.0 compatibility guarantee. It tracks third-party site
// internals (YouTube/innertube/HLS) and will change as those change.
package format

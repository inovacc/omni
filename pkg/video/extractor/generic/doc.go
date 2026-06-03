// Package generic provides a fallback video extractor that scrapes Open Graph
// and HTML5 <video> metadata from arbitrary web pages when no site-specific
// extractor matches the URL.
//
// Experimental: this package's API may change before a stable release and is
// not covered by the v1.0 compatibility guarantee. It tracks third-party site
// internals (YouTube/innertube/HLS) and will change as those change.
package generic

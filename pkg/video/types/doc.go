// Package types defines shared data structures for the video download engine
// including VideoInfo, Format, Fragment, Thumbnail, Subtitle, Chapter,
// and error types. This package exists to break import cycles between
// pkg/video and its sub-packages.
//
// Experimental: this package's API may change before a stable release and is
// not covered by the v1.0 compatibility guarantee. It tracks third-party site
// internals (YouTube/innertube/HLS) and will change as those change. As a
// stable data-only package, it is a candidate for promotion to stable later.
package types

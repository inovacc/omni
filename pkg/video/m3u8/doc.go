// Package m3u8 provides an HLS manifest parser that handles both master
// and media playlists, including segment durations, encryption keys
// (EXT-X-KEY), and variant stream selection.
//
// Experimental: this package's API may change before a stable release and is
// not covered by the v1.0 compatibility guarantee. It tracks third-party site
// internals (YouTube/innertube/HLS) and will change as those change. As a
// self-contained HLS parser, it is a candidate for promotion to stable later.
package m3u8

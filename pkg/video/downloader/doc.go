// Package downloader provides video download engines for HTTP/HTTPS
// and HLS (M3U8) protocols with resume, retry, rate limiting, and
// progress tracking support.
//
// Experimental: this package's API may change before a stable release and is
// not covered by the v1.0 compatibility guarantee. It tracks third-party site
// internals (YouTube/innertube/HLS) and will change as those change.
package downloader

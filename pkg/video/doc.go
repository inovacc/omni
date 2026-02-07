// Package video provides a pure Go video download engine supporting YouTube
// and other video platforms. It is designed as a reusable library that can be
// imported by external Go projects.
//
// The package follows a layered architecture:
//   - types/errors: Core data structures and error types
//   - utils: Filename sanitization, HTML parsing, URL manipulation, duration parsing
//   - nethttp: HTTP client with cookie jar, proxy, retries, custom UA
//   - cache: Filesystem-based cache using XDG paths
//   - jsinterp: JavaScript execution via goja (for YouTube signature decryption)
//   - m3u8: HLS manifest parser
//   - format: Format sorting and selection
//   - downloader: HTTP and HLS download engines with resume support
//   - extractor: Site-specific video metadata extractors
//
// Basic usage:
//
//	client := video.New(video.WithFormat("best"))
//	info, err := client.Extract(ctx, "https://www.youtube.com/watch?v=...")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	err = client.Download(ctx, info)
package video

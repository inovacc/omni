// Package all imports all extractor packages to register them.
// Import this package to ensure all extractors are available:
//
//	import _ "github.com/inovacc/omni/pkg/video/extractor/all"
//
// Experimental: this package's API may change before a stable release and is
// not covered by the v1.0 compatibility guarantee. It tracks third-party site
// internals (YouTube/innertube/HLS) and will change as those change.
package all

import (
	_ "github.com/inovacc/omni/pkg/video/extractor/generic"
	_ "github.com/inovacc/omni/pkg/video/extractor/youtube"
)

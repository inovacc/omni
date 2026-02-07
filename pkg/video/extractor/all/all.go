// Package all imports all extractor packages to register them.
// Import this package to ensure all extractors are available:
//
//	import _ "github.com/inovacc/omni/pkg/video/extractor/all"
package all

import (
	_ "github.com/inovacc/omni/pkg/video/extractor/generic"
	_ "github.com/inovacc/omni/pkg/video/extractor/youtube"
)

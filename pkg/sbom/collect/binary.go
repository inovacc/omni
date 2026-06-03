package collect

import (
	"debug/buildinfo"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/inovacc/omni/pkg/sbom/model"
)

// BinaryFile reads build information embedded in a Go binary at path and builds
// an SBOM describing what actually shipped. A non-Go binary (or unreadable file)
// returns an error the CLI maps to ErrInvalidInput / ErrNotFound.
func BinaryFile(path string) (*model.SBOM, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("read binary %q: %w", path, err)
	}
	bi, err := buildinfo.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("parse build info from %q: %w", path, err)
	}
	return Binary(bi), nil
}

// Binary converts already-parsed build info into an SBOM. The Go toolchain is
// added as a KindToolchain component; replaced modules resolve to the effective
// (replacement) module while preserving the original path/version.
func Binary(bi *debug.BuildInfo) *model.SBOM {
	sb := &model.SBOM{
		Name: lastPathElement(bi.Main.Path),
		Root: model.Component{Path: bi.Main.Path, Version: bi.Main.Version, Kind: model.KindRoot},
	}
	sb.Components = append(sb.Components, model.Component{
		Path: "std", Version: bi.GoVersion, Kind: model.KindToolchain,
	})
	for _, d := range bi.Deps {
		c := model.Component{Path: d.Path, Version: d.Version, Kind: model.KindLibrary}
		if d.Replace != nil {
			c.OriginalPath, c.OriginalVersion = d.Path, d.Version
			c.Path, c.Version = d.Replace.Path, d.Replace.Version
		}
		sb.Components = append(sb.Components, c)
	}
	return sb
}

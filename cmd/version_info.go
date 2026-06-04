package cmd

import "runtime/debug"

// version is overridable at build time via -ldflags "-X github.com/inovacc/omni/cmd.version=vX.Y.Z".
// When unset, rootVersion falls back to the VCS stamp embedded by -buildvcs=true.
var version = ""

// rootVersion returns the omni version string for --version, SBOM tool fields,
// and attestation builder.version. It prefers the ldflags value, then the
// buildvcs main-module version, then "(devel)".
func rootVersion() string {
	if version != "" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "(devel)"
}

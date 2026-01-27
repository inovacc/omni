package cli

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

// UnameOptions configures the uname command behavior
type UnameOptions struct {
	All              bool // -a: print all information
	KernelName       bool // -s: print the kernel name
	NodeName         bool // -n: print the network node hostname
	KernelRelease    bool // -r: print the kernel release
	KernelVersion    bool // -v: print the kernel version
	Machine          bool // -m: print the machine hardware name
	Processor        bool // -p: print the processor type
	HardwarePlatform bool // -i: print the hardware platform
	OperatingSystem  bool // -o: print the operating system
}

// UnameInfo contains system information
type UnameInfo struct {
	KernelName       string `json:"kernelName"`
	NodeName         string `json:"nodeName"`
	KernelRelease    string `json:"kernelRelease"`
	KernelVersion    string `json:"kernelVersion"`
	Machine          string `json:"machine"`
	Processor        string `json:"processor"`
	HardwarePlatform string `json:"hardwarePlatform"`
	OperatingSystem  string `json:"operatingSystem"`
}

// RunUname prints system information
func RunUname(w io.Writer, opts UnameOptions) error {
	info := GetUnameInfo()

	// If no flags specified, default to -s (kernel name)
	if !opts.All && !opts.KernelName && !opts.NodeName && !opts.KernelRelease &&
		!opts.KernelVersion && !opts.Machine && !opts.Processor &&
		!opts.HardwarePlatform && !opts.OperatingSystem {
		opts.KernelName = true
	}

	// If -a, enable all flags
	if opts.All {
		opts.KernelName = true
		opts.NodeName = true
		opts.KernelRelease = true
		opts.KernelVersion = true
		opts.Machine = true
		opts.Processor = true
		opts.HardwarePlatform = true
		opts.OperatingSystem = true
	}

	var parts []string

	if opts.KernelName {
		parts = append(parts, info.KernelName)
	}
	if opts.NodeName {
		parts = append(parts, info.NodeName)
	}
	if opts.KernelRelease {
		parts = append(parts, info.KernelRelease)
	}
	if opts.KernelVersion {
		parts = append(parts, info.KernelVersion)
	}
	if opts.Machine {
		parts = append(parts, info.Machine)
	}
	if opts.Processor {
		parts = append(parts, info.Processor)
	}
	if opts.HardwarePlatform {
		parts = append(parts, info.HardwarePlatform)
	}
	if opts.OperatingSystem {
		parts = append(parts, info.OperatingSystem)
	}

	_, _ = fmt.Fprintln(w, strings.Join(parts, " "))
	return nil
}

// GetUnameInfo returns system information
func GetUnameInfo() UnameInfo {
	hostname, _ := os.Hostname()

	// Map Go's GOOS to more traditional kernel names
	kernelName := mapKernelName(runtime.GOOS)

	// Map Go's GOARCH to machine architecture names
	machine := mapMachine(runtime.GOARCH)

	return UnameInfo{
		KernelName:       kernelName,
		NodeName:         hostname,
		KernelRelease:    getKernelRelease(),
		KernelVersion:    getKernelVersion(),
		Machine:          machine,
		Processor:        machine, // Often same as machine
		HardwarePlatform: machine, // Often same as machine
		OperatingSystem:  mapOperatingSystem(runtime.GOOS),
	}
}

// mapKernelName maps GOOS to traditional kernel names
func mapKernelName(goos string) string {
	switch goos {
	case "linux":
		return "Linux"
	case "darwin":
		return "Darwin"
	case "windows":
		return "Windows_NT"
	case "freebsd":
		return "FreeBSD"
	case "openbsd":
		return "OpenBSD"
	case "netbsd":
		return "NetBSD"
	case "dragonfly":
		return "DragonFly"
	case "solaris":
		return "SunOS"
	case "aix":
		return "AIX"
	default:
		return strings.Title(goos)
	}
}

// mapMachine maps GOARCH to machine architecture names
func mapMachine(goarch string) string {
	switch goarch {
	case "amd64":
		return "x86_64"
	case "386":
		return "i686"
	case "arm":
		return "armv7l"
	case "arm64":
		return "aarch64"
	case "ppc64":
		return "ppc64"
	case "ppc64le":
		return "ppc64le"
	case "mips":
		return "mips"
	case "mips64":
		return "mips64"
	case "riscv64":
		return "riscv64"
	case "s390x":
		return "s390x"
	default:
		return goarch
	}
}

// mapOperatingSystem maps GOOS to OS name
func mapOperatingSystem(goos string) string {
	switch goos {
	case "linux":
		return "GNU/Linux"
	case "darwin":
		return "Darwin"
	case "windows":
		return "Windows"
	case "freebsd", "openbsd", "netbsd", "dragonfly":
		return "BSD"
	default:
		return goos
	}
}

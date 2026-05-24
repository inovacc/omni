// Package gopsagent is an embeddable runtime-introspection agent that Go
// programs can start with one call to Listen(). It exposes opcodes for stack
// dump, GC, memstats, runtime snapshot, CPU/heap profiles, runtime traces,
// and a streaming snapshot feed — all over a loopback TCP socket protected
// by an optional HMAC challenge.
//
// Typical usage in a Go program:
//
//	import "github.com/inovacc/omni/pkg/gopsagent"
//
//	func main() {
//	    a := gopsagent.New(gopsagent.Options{})
//	    if err := a.Listen(); err != nil { log.Fatal(err) }
//	    defer a.Close()
//	    // ... rest of your program
//	}
//
// The `omni gops` CLI is the matching client (omni gops agent-cmd,
// omni gops trace, omni gops profile, omni gops stream).
//
// Adapted from github.com/inovacc/gops (MIT) — see THIRD_PARTY_LICENSES/gops-MIT.txt.
package gopsagent

// Opcode constants — single-byte wire protocol identifiers.
const (
	OpStack           byte = 0x1
	OpGC              byte = 0x2
	OpMemStats        byte = 0x3
	OpVersion         byte = 0x4
	OpHeapProfile     byte = 0x6
	OpCPUProfile      byte = 0x7
	OpStats           byte = 0x8
	OpTrace           byte = 0x9
	OpSetGCPercent    byte = 0x10
	OpRuntimeSnapshot byte = 0x80
	OpMetricsStream   byte = 0x81
	OpAuthChallenge   byte = 0x82
	OpShutdown        byte = 0x83
)

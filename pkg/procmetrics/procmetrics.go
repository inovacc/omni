// Package procmetrics collects external (kernel-observable) metrics for a
// process via gopsutil. It is intentionally agent-free — runtime fields
// (Goroutines, HeapAlloc, GCCount) remain zero unless a caller merges an
// agent snapshot in afterwards.
//
// Adapted from github.com/inovacc/gops (MIT) — see THIRD_PARTY_LICENSES/gops-MIT.txt.
//
// Experimental: this package's API may change before a stable release and is
// not covered by the v1.0 compatibility guarantee.
package procmetrics

import (
	"context"

	"github.com/shirou/gopsutil/v3/process"
)

// Metrics is one snapshot of a process. CPU/Mem/Disk/FD come from the kernel
// via gopsutil; Go-runtime fields (Goroutines, HeapAlloc, GCCount) are only
// populated when a caller merges data from an embedded gops-style agent.
type Metrics struct {
	PID            int32   `json:"pid"`
	CPUPercent     float64 `json:"cpu_percent"`
	MemRSS         uint64  `json:"mem_rss"`
	MemVMS         uint64  `json:"mem_vms"`
	DiskReadBytes  uint64  `json:"disk_read_bytes"`
	DiskWriteBytes uint64  `json:"disk_write_bytes"`
	NetRxBytes     uint64  `json:"net_rx_bytes,omitempty"`
	NetTxBytes     uint64  `json:"net_tx_bytes,omitempty"`
	OpenFDs        int32   `json:"open_fds"`
	Goroutines     int     `json:"goroutines,omitempty"`
	HeapAlloc      uint64  `json:"heap_alloc,omitempty"`
	GCCount        uint32  `json:"gc_count,omitempty"`
}

// Collector gathers external metrics. The zero value is ready to use.
type Collector struct{}

// NewCollector returns a ready-to-use Collector.
func NewCollector() *Collector { return &Collector{} }

// Collect fills the external fields of Metrics for the given PID. Each
// gopsutil call is tolerant of partial failures — fields that cannot be
// read are left at their zero value rather than failing the whole call.
func (c *Collector) Collect(ctx context.Context, pid int32) (Metrics, error) {
	p, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return Metrics{}, err
	}
	m := Metrics{PID: pid}
	if v, err := p.CPUPercentWithContext(ctx); err == nil {
		m.CPUPercent = v
	}
	if mi, err := p.MemoryInfoWithContext(ctx); err == nil && mi != nil {
		m.MemRSS = mi.RSS
		m.MemVMS = mi.VMS
	}
	if io, err := p.IOCountersWithContext(ctx); err == nil && io != nil {
		m.DiskReadBytes = io.ReadBytes
		m.DiskWriteBytes = io.WriteBytes
	}
	if n, err := p.NumFDsWithContext(ctx); err == nil {
		m.OpenFDs = n
	}
	return m, nil
}

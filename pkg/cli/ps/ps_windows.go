//go:build windows

package ps

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	modKernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First           = modKernel32.NewProc("Process32FirstW")
	procProcess32Next            = modKernel32.NewProc("Process32NextW")
	procCloseHandle              = modKernel32.NewProc("CloseHandle")
)

const (
	th32csSnapprocess = 0x00000002
	maxPath           = 260
)

type processEntry32 struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [maxPath]uint16
}

// GetProcessList returns a list of running processes on Windows
func GetProcessList(opts Options) ([]Info, error) {
	var processes []Info

	// Create snapshot of processes
	handle, _, err := procCreateToolhelp32Snapshot.Call(
		uintptr(th32csSnapprocess),
		0,
	)
	if handle == uintptr(syscall.InvalidHandle) {
		return nil, fmt.Errorf("CreateToolhelp32Snapshot failed: %w", err)
	}

	defer func() {
		_, _, _ = procCloseHandle.Call(handle)
	}()

	var entry processEntry32

	entry.Size = uint32(unsafe.Sizeof(entry))

	// Get first process
	ret, _, _ := procProcess32First.Call(handle, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return nil, fmt.Errorf("Process32First failed")
	}

	currentPID := uint32(os.Getpid())

	for {
		// Filter by PID if specified
		if opts.Pid > 0 && int(entry.ProcessID) != opts.Pid {
			goto next
		}

		// If not showing all, only show current process tree
		if !opts.All && !opts.Aux {
			// Simple filter: show only processes with same parent or current process
			if entry.ProcessID != currentPID && entry.ParentProcessID != currentPID {
				goto next
			}
		}

		{
			proc := Info{
				PID:     int(entry.ProcessID),
				PPID:    int(entry.ParentProcessID),
				Command: syscall.UTF16ToString(entry.ExeFile[:]),
				TTY:     "?",
				Stat:    "R",
				Time:    "0:00",
				User:    "SYSTEM", // Would need more API calls to get actual user
			}
			processes = append(processes, proc)
		}

	next:
		// Get next process
		entry.Size = uint32(unsafe.Sizeof(entry))

		ret, _, _ = procProcess32Next.Call(handle, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	return processes, nil
}

//go:build unix

package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// GetProcessList returns a list of running processes on Unix systems
func GetProcessList(opts PsOptions) ([]ProcessInfo, error) {
	var processes []ProcessInfo

	// Read /proc directory
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("cannot read /proc: %w", err)
	}

	currentUID := os.Getuid()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name is a number (PID)
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		// If -p specified, only show that process
		if opts.Pid > 0 && pid != opts.Pid {
			continue
		}

		proc, err := readProcInfo(pid)
		if err != nil {
			continue // Process may have exited
		}

		// Filter by user if not showing all
		if !opts.All && !opts.Aux && proc.UID != currentUID {
			continue
		}

		// Filter by specified user
		if opts.User != "" && proc.User != opts.User {
			continue
		}

		processes = append(processes, proc)
	}

	return processes, nil
}

func readProcInfo(pid int) (ProcessInfo, error) {
	proc := ProcessInfo{PID: pid}
	procPath := filepath.Join("/proc", strconv.Itoa(pid))

	// Read stat file
	statPath := filepath.Join(procPath, "stat")
	statData, err := os.ReadFile(statPath)
	if err != nil {
		return proc, err
	}

	// Parse stat - format: pid (comm) state ppid pgrp session tty_nr ...
	statStr := string(statData)

	// Find command name between parentheses
	start := strings.Index(statStr, "(")
	end := strings.LastIndex(statStr, ")")
	if start == -1 || end == -1 {
		return proc, fmt.Errorf("invalid stat format")
	}

	proc.Command = statStr[start+1 : end]
	fields := strings.Fields(statStr[end+2:])

	if len(fields) >= 2 {
		proc.Stat = fields[0]
		proc.PPID, _ = strconv.Atoi(fields[1])
	}

	if len(fields) >= 5 {
		ttyNr, _ := strconv.Atoi(fields[4])
		proc.TTY = formatTTY(ttyNr)
	}

	if len(fields) >= 12 {
		// CPU time: utime + stime (fields 11 and 12, 0-indexed from after state)
		utime, _ := strconv.ParseInt(fields[11], 10, 64)
		stime, _ := strconv.ParseInt(fields[12], 10, 64)
		totalTicks := utime + stime
		proc.Time = formatCPUTime(totalTicks)
	}

	if len(fields) >= 21 {
		// VSZ is in pages, convert to KB (assuming 4KB pages)
		vsize, _ := strconv.ParseInt(fields[20], 10, 64)
		proc.VSZ = vsize / 1024
	}

	// Read statm for RSS
	statmPath := filepath.Join(procPath, "statm")
	if statmData, err := os.ReadFile(statmPath); err == nil {
		statmFields := strings.Fields(string(statmData))
		if len(statmFields) >= 2 {
			rssPages, _ := strconv.ParseInt(statmFields[1], 10, 64)
			proc.RSS = rssPages * 4 // 4KB pages to KB
		}
	}

	// Read status for UID
	statusPath := filepath.Join(procPath, "status")
	if statusFile, err := os.Open(statusPath); err == nil {
		scanner := bufio.NewScanner(statusFile)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Uid:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					proc.UID, _ = strconv.Atoi(fields[1])
				}
				break
			}
		}
		_ = statusFile.Close()
	}

	// Get username from UID
	if u, err := user.LookupId(strconv.Itoa(proc.UID)); err == nil {
		proc.User = u.Username
	} else {
		proc.User = strconv.Itoa(proc.UID)
	}

	// Read cmdline for full command
	cmdlinePath := filepath.Join(procPath, "cmdline")
	if cmdlineData, err := os.ReadFile(cmdlinePath); err == nil && len(cmdlineData) > 0 {
		// Replace null bytes with spaces
		cmdline := strings.ReplaceAll(string(cmdlineData), "\x00", " ")
		cmdline = strings.TrimSpace(cmdline)
		if cmdline != "" {
			proc.Command = cmdline
		}
	}

	// Calculate CPU and MEM percentages (simplified)
	proc.CPU = 0.0 // Would need to sample over time for accurate value
	proc.MEM = 0.0 // Would need total memory for accurate value

	// Start time (simplified)
	proc.Start = "?"

	return proc, nil
}

func formatTTY(ttyNr int) string {
	if ttyNr == 0 {
		return "?"
	}
	major := (ttyNr >> 8) & 0xff
	minor := ttyNr & 0xff

	switch major {
	case 4:
		return fmt.Sprintf("tty%d", minor)
	case 136:
		return fmt.Sprintf("pts/%d", minor)
	default:
		return fmt.Sprintf("%d/%d", major, minor)
	}
}

func formatCPUTime(ticks int64) string {
	// Assuming 100 ticks per second (CLK_TCK)
	totalSeconds := ticks / 100
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

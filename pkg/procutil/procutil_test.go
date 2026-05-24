package procutil

import (
	"context"
	"os"
	"testing"
)

func TestListAll_FindsSelf(t *testing.T) {
	procs, err := List(context.Background(), ListOptions{IncludeSelf: true})
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	selfPID := int32(os.Getpid())
	for _, p := range procs {
		if p.PID == selfPID {
			if p.Runtime != RuntimeGo {
				t.Errorf("self process classified as %q, want %q", p.Runtime, RuntimeGo)
			}
			if p.GoVersion == "" {
				t.Error("self GoVersion should be populated")
			}
			return
		}
	}
	t.Fatalf("self PID %d not found in ListAll(IncludeSelf=true) result of %d procs", selfPID, len(procs))
}

func TestList_ExcludesSelfByDefault(t *testing.T) {
	procs, err := List(context.Background(), ListOptions{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	selfPID := int32(os.Getpid())
	for _, p := range procs {
		if p.PID == selfPID {
			t.Errorf("self PID %d leaked into default List() result", selfPID)
		}
	}
}

func TestListByRuntime_Filter(t *testing.T) {
	procs, err := ListByRuntime(context.Background(), RuntimeGo)
	if err != nil {
		t.Fatalf("ListByRuntime(Go): %v", err)
	}
	for _, p := range procs {
		if p.Runtime != RuntimeGo {
			t.Errorf("ListByRuntime(Go) returned non-Go process: pid=%d runtime=%q", p.PID, p.Runtime)
		}
	}
}

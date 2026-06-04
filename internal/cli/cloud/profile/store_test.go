package profile

import (
	"os"
	"runtime"
	"testing"
)

// TestSaveCredentials_ResetsLoosePermsOnPreExistingFile is a regression test for
// finding [savecreds-nonatomic-perm]: a pre-planted credentials file with loose
// 0644 perms must end up 0600 after SaveCredentials, since os.WriteFile does NOT
// reset perms on an already-existing file (CWE-367/732). The fix routes
// SaveCredentials through atomicWriteFile, which creates a fresh 0600 temp file
// and renames it over the path regardless of any pre-existing mode.
//
// Perm bits are POSIX-specific; skip on Windows where chmod is effectively a
// no-op for these bits.
func TestSaveCredentials_ResetsLoosePermsOnPreExistingFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping: POSIX file permission semantics not enforced on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "omni-store-perm-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	store := NewFileStore(tmpDir)

	const (
		provider = Provider("aws")
		name     = "default"
	)

	// Ensure the directory exists, then pre-plant a credentials file with loose
	// 0644 perms to simulate an attacker-planted or legacy world-readable file.
	if err := store.ensureDir(provider); err != nil {
		t.Fatalf("ensureDir failed: %v", err)
	}

	path := store.credentialsPath(provider, name)
	if err := os.WriteFile(path, []byte("old-loose-content"), 0644); err != nil {
		t.Fatalf("pre-creating loose cred file failed: %v", err)
	}

	// Confirm the pre-condition: file really is 0644.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat pre-created file failed: %v", err)
	}

	if got := info.Mode().Perm(); got != 0644 {
		t.Fatalf("precondition: expected pre-created file mode 0644, got %o", got)
	}

	// Save credentials over the pre-existing loose file.
	if err := store.SaveCredentials(provider, name, []byte("secret-encrypted-bytes")); err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	info, err = os.Stat(path)
	if err != nil {
		t.Fatalf("stat after SaveCredentials failed: %v", err)
	}

	if got := info.Mode().Perm(); got != credentialPerms {
		t.Errorf("SaveCredentials did not reset perms: expected %o, got %o", os.FileMode(credentialPerms), got)
	}

	// Sanity: content should be the new credentials, not the old loose content.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved creds failed: %v", err)
	}

	if string(data) != "secret-encrypted-bytes" {
		t.Errorf("unexpected credentials content: %q", string(data))
	}
}

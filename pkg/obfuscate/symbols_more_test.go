package obfuscate

import (
	"debug/elf"
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// tempELFDir returns a temp dir for ELF fixtures with cleanup that tolerates the
// Windows handle race: debug/elf.NewFile(f) does NOT take ownership of the
// underlying *os.File on Windows (its closer is nil), so hasGoSymbols leaves the
// fixture fd open until the *os.File is finalized. We force finalization with a
// GC before removing the dir so cleanup does not fail on "file in use".
func tempELFDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "obfuscate-elf-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() {
		runtime.GC()
		runtime.GC() // second pass: run finalizers queued by the first
		_ = os.RemoveAll(dir)
	})
	return dir
}

// buildMinimalELF writes a tiny but valid little-endian ELF64 object to path,
// containing a NULL section, a .shstrtab, and one section whose name is
// sectionName. It is just enough for debug/elf.NewFile to enumerate sections,
// which is all hasGoSymbols needs. ELF parsing is byte-based and host-agnostic,
// so this exercises the ELF branch even when running on Windows/macOS.
func buildMinimalELF(t *testing.T, path, sectionName string) {
	t.Helper()

	// Section header string table: "\0.shstrtab\0<sectionName>\0".
	shstr := []byte{0}
	shstrtabOff := len(shstr)
	shstr = append(shstr, ".shstrtab"...)
	shstr = append(shstr, 0)
	nameOff := len(shstr)
	shstr = append(shstr, sectionName...)
	shstr = append(shstr, 0)

	const ehSize = 64 // Elf64 header
	const shSize = 64 // Elf64 section header entry
	numSections := 3  // NULL, .shstrtab, target
	shoff := uint64(ehSize)
	shstrDataOff := shoff + uint64(numSections*shSize)

	buf := make([]byte, shstrDataOff+uint64(len(shstr)))

	// --- ELF header ---
	copy(buf[0:], []byte{0x7f, 'E', 'L', 'F'})
	buf[4] = 2 // ELFCLASS64
	buf[5] = 1 // ELFDATA2LSB (little-endian)
	buf[6] = 1 // EV_CURRENT
	le := binary.LittleEndian
	le.PutUint16(buf[16:], uint16(elf.ET_EXEC))
	le.PutUint16(buf[18:], uint16(elf.EM_X86_64))
	le.PutUint32(buf[20:], 1) // version
	le.PutUint64(buf[40:], shoff)
	le.PutUint16(buf[58:], shSize)              // e_shentsize
	le.PutUint16(buf[60:], uint16(numSections)) // e_shnum
	le.PutUint16(buf[62:], 1)                   // e_shstrndx -> section 1 (.shstrtab)

	putSH := func(idx int, name uint32, typ elf.SectionType, off, size uint64) {
		base := int(shoff) + idx*shSize
		le.PutUint32(buf[base+0:], name)
		le.PutUint32(buf[base+4:], uint32(typ))
		le.PutUint64(buf[base+24:], off)  // sh_offset
		le.PutUint64(buf[base+32:], size) // sh_size
	}
	// Section 0: NULL.
	putSH(0, 0, elf.SHT_NULL, 0, 0)
	// Section 1: .shstrtab.
	putSH(1, uint32(shstrtabOff), elf.SHT_STRTAB, shstrDataOff, uint64(len(shstr)))
	// Section 2: the target section (PROGBITS, empty payload).
	putSH(2, uint32(nameOff), elf.SHT_PROGBITS, shstrDataOff, 0)

	copy(buf[shstrDataOff:], shstr)

	if err := os.WriteFile(path, buf, 0o600); err != nil {
		t.Fatalf("write ELF fixture: %v", err)
	}
}

// TestHasGoSymbolsELFWithPclntab covers the ELF symbol-found return-true branch.
func TestHasGoSymbolsELFWithPclntab(t *testing.T) {
	path := filepath.Join(tempELFDir(t), "go.elf")
	buildMinimalELF(t, path, ".gopclntab")

	ok, err := hasGoSymbols(path)
	if err != nil {
		t.Fatalf("hasGoSymbols: %v", err)
	}
	if !ok {
		t.Error("hasGoSymbols(.gopclntab ELF) = false, want true")
	}
}

// TestHasGoSymbolsELFWithoutPclntab covers the ELF return-false branch via a
// valid ELF whose only named section is unrelated.
func TestHasGoSymbolsELFWithoutPclntab(t *testing.T) {
	path := filepath.Join(tempELFDir(t), "plain.elf")
	buildMinimalELF(t, path, ".text")

	ok, err := hasGoSymbols(path)
	if err != nil {
		t.Fatalf("hasGoSymbols: %v", err)
	}
	if ok {
		t.Error("hasGoSymbols(.text-only ELF) = true, want false")
	}
}

// TestHasGoSymbolsNonBinary covers the final fallthrough (not ELF/PE/macho).
func TestHasGoSymbolsNonBinary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plain.txt")
	if err := os.WriteFile(path, []byte("just some text, not a binary at all\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	ok, err := hasGoSymbols(path)
	if err != nil {
		t.Fatalf("hasGoSymbols(text): %v", err)
	}
	if ok {
		t.Error("hasGoSymbols(text file) = true, want false")
	}
}

// TestHasGoSymbolsMissingFile covers the os.Open error path.
func TestHasGoSymbolsMissingFile(t *testing.T) {
	if _, err := hasGoSymbols(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Error("hasGoSymbols(missing) = nil error, want error")
	}
}

// TestDetectELFFixtureVerdicts drives Detect over the crafted fixtures: an ELF
// with .gopclntab but no buildinfo is symbols-only -> suspect-stripped; a plain
// text file is not_go.
func TestDetectELFFixtureVerdicts(t *testing.T) {
	dir := tempELFDir(t)

	goELF := filepath.Join(dir, "withsym.elf")
	buildMinimalELF(t, goELF, ".gopclntab")
	v, err := Detect(goELF)
	if err != nil {
		t.Fatalf("Detect(symbols ELF): %v", err)
	}
	if !v.SymbolsFound {
		t.Errorf("SymbolsFound = false, want true (verdict=%s)", v.Verdict)
	}
	// No buildinfo in our synthetic ELF, but symbols present -> suspect-stripped.
	if v.Verdict != VerdictSuspectStripped {
		t.Errorf("verdict = %s, want %s", v.Verdict, VerdictSuspectStripped)
	}

	txt := filepath.Join(dir, "notgo.txt")
	if err := os.WriteFile(txt, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	v2, err := Detect(txt)
	if err != nil {
		t.Fatalf("Detect(text): %v", err)
	}
	if v2.Verdict != VerdictNotGo {
		t.Errorf("verdict = %s, want %s", v2.Verdict, VerdictNotGo)
	}
}

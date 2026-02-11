package note

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/inovacc/omni/pkg/userdirs"
)

const defaultFileName = "omni-notes.json"

// Options configures note command behavior.
type Options struct {
	File  string
	List  bool
	JSON  bool
	Limit int
}

// Entry is a single note item.
type Entry struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

// Store is the JSON structure persisted on disk.
type Store struct {
	Notes []Entry `json:"notes"`
}

// RunNote runs the note command.
func RunNote(w io.Writer, args []string, opts Options) error {
	path, err := resolveNotesPath(opts.File)
	if err != nil {
		return fmt.Errorf("note: %w", err)
	}

	if opts.List {
		return runList(w, path, opts)
	}

	text := strings.TrimSpace(strings.Join(args, " "))
	if text == "" {
		return fmt.Errorf("note: text is required (example: omni note \"buy milk\")")
	}

	entry, err := add(path, text)
	if err != nil {
		return err
	}

	if opts.JSON {
		data, err := json.MarshalIndent(entry, "", "  ")
		if err != nil {
			return fmt.Errorf("note: marshal: %w", err)
		}

		_, _ = fmt.Fprintln(w, string(data))
		return nil
	}

	_, _ = fmt.Fprintf(w, "Saved note %s\n", entry.ID)
	_, _ = fmt.Fprintf(w, "File: %s\n", path)

	return nil
}

// RunRemove deletes a note entry by 1-based index or note ID.
func RunRemove(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("note: remove target is required (index or note id)")
	}

	target := strings.TrimSpace(args[0])
	if target == "" {
		return fmt.Errorf("note: remove target is required (index or note id)")
	}

	path, err := resolveNotesPath(opts.File)
	if err != nil {
		return fmt.Errorf("note: %w", err)
	}

	removed, index, err := remove(path, target)
	if err != nil {
		return err
	}

	if opts.JSON {
		data, err := json.MarshalIndent(removed, "", "  ")
		if err != nil {
			return fmt.Errorf("note: marshal: %w", err)
		}

		_, _ = fmt.Fprintln(w, string(data))
		return nil
	}

	_, _ = fmt.Fprintf(w, "Removed note #%d (%s)\n", index, removed.ID)

	return nil
}

func runList(w io.Writer, path string, opts Options) error {
	store, err := load(path)
	if err != nil {
		return err
	}

	notes := store.Notes
	if opts.Limit > 0 && opts.Limit < len(notes) {
		notes = notes[len(notes)-opts.Limit:]
	}

	if opts.JSON {
		data, err := json.MarshalIndent(Store{Notes: notes}, "", "  ")
		if err != nil {
			return fmt.Errorf("note: marshal: %w", err)
		}

		_, _ = fmt.Fprintln(w, string(data))
		return nil
	}

	if len(notes) == 0 {
		_, _ = fmt.Fprintln(w, "No notes yet.")
		return nil
	}

	for i, n := range notes {
		_, _ = fmt.Fprintf(w, "%d. [%s] %s (id=%s)\n", i+1, n.CreatedAt, n.Text, n.ID)
	}

	return nil
}

func remove(path, target string) (Entry, int, error) {
	store, err := load(path)
	if err != nil {
		return Entry{}, 0, err
	}

	if len(store.Notes) == 0 {
		return Entry{}, 0, fmt.Errorf("note: no notes to remove")
	}

	var idx int = -1
	parsedIndex := false
	parsedIndexValue := 0

	if n, convErr := strconv.Atoi(target); convErr == nil {
		parsedIndex = true
		parsedIndexValue = n
		if n >= 1 && n <= len(store.Notes) {
			idx = n - 1
		}
	}

	if idx == -1 {
		for i, entry := range store.Notes {
			if entry.ID == target {
				idx = i
				break
			}
		}
	}

	if idx == -1 {
		if parsedIndex {
			return Entry{}, 0, fmt.Errorf("note: index out of range: %d", parsedIndexValue)
		}

		return Entry{}, 0, fmt.Errorf("note: entry not found: %s", target)
	}

	removed := store.Notes[idx]
	store.Notes = append(store.Notes[:idx], store.Notes[idx+1:]...)

	if err := save(path, store); err != nil {
		return Entry{}, 0, err
	}

	return removed, idx + 1, nil
}

func add(path, text string) (Entry, error) {
	store, err := load(path)
	if err != nil {
		return Entry{}, err
	}

	now := time.Now().UTC()
	entry := Entry{
		ID:        strconv.FormatInt(now.UnixNano(), 10),
		Text:      text,
		CreatedAt: now.Format(time.RFC3339),
	}

	store.Notes = append(store.Notes, entry)

	if err := save(path, store); err != nil {
		return Entry{}, err
	}

	return entry, nil
}

func load(path string) (Store, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Store{}, nil
		}

		return Store{}, fmt.Errorf("note: read %s: %w", path, err)
	}

	var store Store
	if len(data) == 0 {
		return store, nil
	}

	if err := json.Unmarshal(data, &store); err != nil {
		return Store{}, fmt.Errorf("note: parse %s: %w", path, err)
	}

	return store, nil
}

func save(path string, store Store) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("note: create directory: %w", err)
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("note: marshal: %w", err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("note: write %s: %w", path, err)
	}

	return nil
}

func resolveNotesPath(custom string) (string, error) {
	if custom != "" {
		return custom, nil
	}

	docs, err := userdirs.DocumentsDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(docs, defaultFileName), nil
}

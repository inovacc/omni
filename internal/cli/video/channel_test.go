package video

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/pkg/video/types"
)

func TestRunChannel_NoArgs(t *testing.T) {
	var buf bytes.Buffer
	err := RunChannel(&buf, nil, Options{})
	if err == nil {
		t.Fatal("expected error for no args")
	}

	if err.Error() != "video channel: URL is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChannelDB_InitAndSchema(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := initChannelDB(dbPath)
	if err != nil {
		t.Fatalf("initChannelDB: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Verify tables exist by querying them.
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='channel'").Scan(&tableName)
	if err != nil {
		t.Fatalf("channel table not found: %v", err)
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='videos'").Scan(&tableName)
	if err != nil {
		t.Fatalf("videos table not found: %v", err)
	}
}

func TestGetDownloadedVideoIDs_Empty(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := initChannelDB(dbPath)
	if err != nil {
		t.Fatalf("initChannelDB: %v", err)
	}
	defer func() { _ = db.Close() }()

	seen, err := getDownloadedVideoIDs(db)
	if err != nil {
		t.Fatalf("getDownloadedVideoIDs: %v", err)
	}

	if len(seen) != 0 {
		t.Errorf("expected empty set, got %d entries", len(seen))
	}
}

func TestInsertVideoRecord_Incremental(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := initChannelDB(dbPath)
	if err != nil {
		t.Fatalf("initChannelDB: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Insert channel first (foreign key).
	channelInfo := &types.VideoInfo{
		ID:         "UCtest123",
		Title:      "Test Channel",
		ChannelURL: "https://www.youtube.com/channel/UCtest123",
	}

	if err := upsertChannel(db, channelInfo); err != nil {
		t.Fatalf("upsertChannel: %v", err)
	}

	// Insert a video.
	videoInfo := &types.VideoInfo{
		ID:         "vid001",
		Title:      "Test Video",
		WebpageURL: "https://www.youtube.com/watch?v=vid001",
		Duration:   120,
		UploadDate: "20240101",
	}

	if err := insertVideoRecord(db, videoInfo, "UCtest123"); err != nil {
		t.Fatalf("insertVideoRecord: %v", err)
	}

	// Verify it shows up in the seen set.
	seen, err := getDownloadedVideoIDs(db)
	if err != nil {
		t.Fatalf("getDownloadedVideoIDs: %v", err)
	}

	if _, exists := seen["vid001"]; !exists {
		t.Error("vid001 not found in seen set after insert")
	}

	if len(seen) != 1 {
		t.Errorf("expected 1 entry in seen set, got %d", len(seen))
	}

	// Insert same video again (should be ignored, not error).
	if err := insertVideoRecord(db, videoInfo, "UCtest123"); err != nil {
		t.Fatalf("insertVideoRecord duplicate: %v", err)
	}

	seen2, err := getDownloadedVideoIDs(db)
	if err != nil {
		t.Fatalf("getDownloadedVideoIDs after dup: %v", err)
	}

	if len(seen2) != 1 {
		t.Errorf("expected 1 entry after duplicate insert, got %d", len(seen2))
	}
}

func TestUpsertChannel_Update(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := initChannelDB(dbPath)
	if err != nil {
		t.Fatalf("initChannelDB: %v", err)
	}
	defer func() { _ = db.Close() }()

	info := &types.VideoInfo{
		ID:          "UCtest",
		Title:       "Old Name",
		ChannelURL:  "https://www.youtube.com/channel/UCtest",
		Description: "old desc",
	}

	if err := upsertChannel(db, info); err != nil {
		t.Fatalf("first upsertChannel: %v", err)
	}

	// Update.
	info.Title = "New Name"
	info.Description = "new desc"

	if err := upsertChannel(db, info); err != nil {
		t.Fatalf("second upsertChannel: %v", err)
	}

	var name, desc string
	err = db.QueryRow("SELECT name, description FROM channel WHERE id = ?", "UCtest").Scan(&name, &desc)
	if err != nil {
		t.Fatalf("query updated channel: %v", err)
	}

	if name != "New Name" {
		t.Errorf("name = %q, want %q", name, "New Name")
	}

	if desc != "new desc" {
		t.Errorf("description = %q, want %q", desc, "new desc")
	}
}

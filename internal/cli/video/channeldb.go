package video

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/inovacc/omni/pkg/video/types"
)

const channelSchema = `
CREATE TABLE IF NOT EXISTS channel (
    id               TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    url              TEXT NOT NULL,
    description      TEXT DEFAULT '',
    subscriber_count TEXT DEFAULT '',
    avatar_url       TEXT DEFAULT '',
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS videos (
    id              TEXT PRIMARY KEY,
    channel_id      TEXT NOT NULL REFERENCES channel(id),
    title           TEXT NOT NULL,
    url             TEXT NOT NULL,
    description     TEXT DEFAULT '',
    duration        REAL DEFAULT 0,
    upload_date     TEXT DEFAULT '',
    view_count      INTEGER,
    like_count      INTEGER,
    thumbnail_url   TEXT DEFAULT '',
    format_id       TEXT DEFAULT '',
    resolution      TEXT DEFAULT '',
    filesize        INTEGER,
    downloaded_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

// initChannelDB opens a SQLite database at dbPath, creates the schema, and
// enables WAL mode for concurrent reads.
func initChannelDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("channeldb: open: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("channeldb: WAL mode: %w", err)
	}

	if _, err := db.Exec(channelSchema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("channeldb: create schema: %w", err)
	}

	return db, nil
}

// upsertChannel inserts or updates channel metadata.
func upsertChannel(db *sql.DB, info *types.VideoInfo) error {
	subscriberCount := ""
	avatarURL := ""

	if info.Metadata != nil {
		subscriberCount = info.Metadata["subscriber_count"]
		avatarURL = info.Metadata["avatar_url"]
	}

	channelURL := info.ChannelURL
	if channelURL == "" {
		channelURL = info.WebpageURL
	}

	_, err := db.Exec(`
		INSERT INTO channel (id, name, url, description, subscriber_count, avatar_url)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			url = excluded.url,
			description = excluded.description,
			subscriber_count = excluded.subscriber_count,
			avatar_url = excluded.avatar_url,
			updated_at = CURRENT_TIMESTAMP
	`, info.ID, info.Title, channelURL, info.Description, subscriberCount, avatarURL)
	if err != nil {
		return fmt.Errorf("channeldb: upsert channel: %w", err)
	}

	return nil
}

// getDownloadedVideoIDs returns a set of video IDs that have already been downloaded.
func getDownloadedVideoIDs(db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.Query("SELECT id FROM videos")
	if err != nil {
		return nil, fmt.Errorf("channeldb: query videos: %w", err)
	}
	defer func() { _ = rows.Close() }()

	seen := make(map[string]struct{})
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("channeldb: scan video id: %w", err)
		}

		seen[id] = struct{}{}
	}

	return seen, rows.Err()
}

// insertVideoRecord records a successfully downloaded video in the database.
func insertVideoRecord(db *sql.DB, info *types.VideoInfo, channelID string) error {
	var viewCount, likeCount *int64
	if info.ViewCount != nil {
		viewCount = info.ViewCount
	}

	if info.LikeCount != nil {
		likeCount = info.LikeCount
	}

	thumbnailURL := ""
	if len(info.Thumbnails) > 0 {
		thumbnailURL = info.Thumbnails[0].URL
	}

	_, err := db.Exec(`
		INSERT OR IGNORE INTO videos (id, channel_id, title, url, description, duration, upload_date, view_count, like_count, thumbnail_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, info.ID, channelID, info.Title, info.WebpageURL, info.Description, info.Duration, info.UploadDate, viewCount, likeCount, thumbnailURL)
	if err != nil {
		return fmt.Errorf("channeldb: insert video: %w", err)
	}

	return nil
}

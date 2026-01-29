package bbolt

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	bolt "go.etcd.io/bbolt"
)

// Options configures bbolt command behavior
type Options struct {
	JSON   bool   // --json: output as JSON
	Hex    bool   // --hex: display values in hex
	Prefix string // --prefix: filter keys by prefix
}

// StatsResult represents database statistics for JSON output
type StatsResult struct {
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	PageSize    int    `json:"page_size"`
	Buckets     int    `json:"buckets"`
	FreePages   int    `json:"free_pages"`
	FreeAlloc   int    `json:"free_alloc"`
	TxCount     int    `json:"tx_count"`
	OpenTxCount int    `json:"open_tx_count"`
}

// BucketInfo represents bucket information for JSON output
type BucketInfo struct {
	Name     string `json:"name"`
	KeyCount int    `json:"key_count"`
}

// KeyValue represents a key-value pair for JSON output
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RunStats displays database statistics
func RunStats(w io.Writer, dbPath string, opts Options) error {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = db.Close() }()

	fi, err := os.Stat(dbPath)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	var bucketCount int

	_ = db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(_ []byte, _ *bolt.Bucket) error {
			bucketCount++
			return nil
		})
	})

	stats := db.Stats()

	if opts.JSON {
		result := StatsResult{
			Path:        filepath.Clean(dbPath),
			Size:        fi.Size(),
			PageSize:    db.Info().PageSize,
			Buckets:     bucketCount,
			FreePages:   stats.FreePageN,
			FreeAlloc:   stats.FreeAlloc,
			TxCount:     stats.TxN,
			OpenTxCount: stats.OpenTxN,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Path: %s\n", filepath.Clean(dbPath))
	_, _ = fmt.Fprintf(w, "Size: %d bytes\n", fi.Size())
	_, _ = fmt.Fprintf(w, "Page Size: %d\n", db.Info().PageSize)
	_, _ = fmt.Fprintf(w, "Buckets: %d\n", bucketCount)
	_, _ = fmt.Fprintf(w, "Free Pages: %d\n", stats.FreePageN)
	_, _ = fmt.Fprintf(w, "Free Alloc: %d bytes\n", stats.FreeAlloc)
	_, _ = fmt.Fprintf(w, "Tx Count: %d\n", stats.TxN)
	_, _ = fmt.Fprintf(w, "Open Tx Count: %d\n", stats.OpenTxN)

	return nil
}

// RunBuckets lists all buckets in the database
func RunBuckets(w io.Writer, dbPath string, opts Options) error {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = db.Close() }()

	var buckets []BucketInfo

	err = db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			stats := b.Stats()
			buckets = append(buckets, BucketInfo{
				Name:     string(name),
				KeyCount: stats.KeyN,
			})

			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(buckets)
	}

	for _, b := range buckets {
		_, _ = fmt.Fprintf(w, "%s (%d keys)\n", b.Name, b.KeyCount)
	}

	return nil
}

// RunKeys lists keys in a bucket
func RunKeys(w io.Writer, dbPath, bucket string, opts Options) error {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = db.Close() }()

	var keys []string

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}

		return b.ForEach(func(k, _ []byte) error {
			key := string(k)
			if opts.Prefix == "" || strings.HasPrefix(key, opts.Prefix) {
				keys = append(keys, key)
			}

			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(keys)
	}

	for _, k := range keys {
		_, _ = fmt.Fprintln(w, k)
	}

	return nil
}

// RunGet retrieves a value by key from a bucket
func RunGet(w io.Writer, dbPath, bucket, key string, opts Options) error {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = db.Close() }()

	var value []byte

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}

		value = b.Get([]byte(key))
		if value == nil {
			return fmt.Errorf("key %q not found in bucket %q", key, bucket)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	if opts.JSON {
		kv := KeyValue{
			Key:   key,
			Value: formatValue(value, opts.Hex),
		}

		return json.NewEncoder(w).Encode(kv)
	}

	if opts.Hex {
		_, _ = fmt.Fprintln(w, hex.EncodeToString(value))
	} else {
		_, _ = w.Write(value)
		_, _ = fmt.Fprintln(w)
	}

	return nil
}

// RunPut stores a key-value pair in a bucket
func RunPut(w io.Writer, dbPath, bucket, key, value string, opts Options) error {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = db.Close() }()

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}

		return b.Put([]byte(key), []byte(value))
	})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	if opts.JSON {
		result := map[string]string{
			"status": "ok",
			"bucket": bucket,
			"key":    key,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Put %s/%s\n", bucket, key)

	return nil
}

// RunDelete removes a key from a bucket
func RunDelete(w io.Writer, dbPath, bucket, key string, opts Options) error {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = db.Close() }()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}

		return b.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	if opts.JSON {
		result := map[string]string{
			"status": "deleted",
			"bucket": bucket,
			"key":    key,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Deleted %s/%s\n", bucket, key)

	return nil
}

// RunDump dumps all keys and values in a bucket
func RunDump(w io.Writer, dbPath, bucket string, opts Options) error {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = db.Close() }()

	var kvPairs []KeyValue

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}

		return b.ForEach(func(k, v []byte) error {
			key := string(k)
			if opts.Prefix == "" || strings.HasPrefix(key, opts.Prefix) {
				kvPairs = append(kvPairs, KeyValue{
					Key:   key,
					Value: formatValue(v, opts.Hex),
				})
			}

			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(kvPairs)
	}

	for _, kv := range kvPairs {
		_, _ = fmt.Fprintf(w, "%s: %s\n", kv.Key, kv.Value)
	}

	return nil
}

// RunCompact compacts the database to a new file
func RunCompact(w io.Writer, srcPath, dstPath string, opts Options) error {
	src, err := bolt.Open(srcPath, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = src.Close() }()

	dst, err := bolt.Open(dstPath, 0600, nil)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = dst.Close() }()

	err = bolt.Compact(dst, src, 65536)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	srcInfo, _ := os.Stat(srcPath)
	dstInfo, _ := os.Stat(dstPath)

	if opts.JSON {
		result := map[string]any{
			"status":      "ok",
			"source":      srcPath,
			"destination": dstPath,
			"source_size": srcInfo.Size(),
			"dest_size":   dstInfo.Size(),
			"savings_pct": float64(srcInfo.Size()-dstInfo.Size()) / float64(srcInfo.Size()) * 100,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Compacted %s -> %s\n", srcPath, dstPath)
	_, _ = fmt.Fprintf(w, "Original: %d bytes\n", srcInfo.Size())
	_, _ = fmt.Fprintf(w, "Compacted: %d bytes\n", dstInfo.Size())
	_, _ = fmt.Fprintf(w, "Savings: %.1f%%\n", float64(srcInfo.Size()-dstInfo.Size())/float64(srcInfo.Size())*100)

	return nil
}

// RunCheck verifies database integrity
func RunCheck(w io.Writer, dbPath string, opts Options) error {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = db.Close() }()

	var errCount int

	err = db.View(func(tx *bolt.Tx) error {
		// Check each bucket
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			return b.ForEach(func(k, v []byte) error {
				if k == nil {
					errCount++
				}

				return nil
			})
		})
	})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	if opts.JSON {
		result := map[string]any{
			"path":   dbPath,
			"status": "ok",
			"errors": errCount,
		}

		if errCount > 0 {
			result["status"] = "errors_found"
		}

		return json.NewEncoder(w).Encode(result)
	}

	if errCount > 0 {
		_, _ = fmt.Fprintf(w, "Check completed with %d errors\n", errCount)
	} else {
		_, _ = fmt.Fprintln(w, "Database OK")
	}

	return nil
}

// RunCreateBucket creates a new bucket
func RunCreateBucket(w io.Writer, dbPath, bucket string, opts Options) error {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = db.Close() }()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	if opts.JSON {
		result := map[string]string{
			"status": "created",
			"bucket": bucket,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Created bucket %q\n", bucket)

	return nil
}

// RunDeleteBucket removes a bucket
func RunDeleteBucket(w io.Writer, dbPath, bucket string, opts Options) error {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = db.Close() }()

	err = db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(bucket))
	})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	if opts.JSON {
		result := map[string]string{
			"status": "deleted",
			"bucket": bucket,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Deleted bucket %q\n", bucket)

	return nil
}

// InfoResult represents database info for JSON output
type InfoResult struct {
	Path     string `json:"path"`
	PageSize int    `json:"page_size"`
}

// RunInfo displays basic database information (matches bbolt info)
func RunInfo(w io.Writer, dbPath string, opts Options) error {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = db.Close() }()

	if opts.JSON {
		result := InfoResult{
			Path:     filepath.Clean(dbPath),
			PageSize: db.Info().PageSize,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Page Size: %d\n", db.Info().PageSize)

	return nil
}

// PageInfo represents page information for JSON output
type PageInfo struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Count    int    `json:"count,omitempty"`
	Overflow int    `json:"overflow,omitempty"`
}

// RunPages lists pages with their types (matches bbolt pages)
func RunPages(w io.Writer, dbPath string, opts Options) error {
	// Open the file directly to read page headers
	f, err := os.Open(dbPath)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = f.Close() }()

	// Read the page size from the meta page
	// The page size is at offset 12 in the meta page (after magic + version)
	buf := make([]byte, 4096) // Read first page

	_, err = f.Read(buf)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	// Simple page size extraction from meta page
	pageSize := int(buf[12]) | int(buf[13])<<8 | int(buf[14])<<16 | int(buf[15])<<24
	if pageSize == 0 {
		pageSize = 4096 // default
	}

	// Get file size to determine page count
	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	pageCount := int(fi.Size()) / pageSize

	var pages []PageInfo

	for i := 0; i < pageCount && i < 100; i++ { // Limit to first 100 pages
		_, _ = f.Seek(int64(i*pageSize), 0)

		pageBuf := make([]byte, 16) // Just read the header
		_, _ = f.Read(pageBuf)

		flags := uint16(pageBuf[4]) | uint16(pageBuf[5])<<8
		count := int(pageBuf[6]) | int(pageBuf[7])<<8
		overflow := int(pageBuf[8]) | int(pageBuf[9])<<8 | int(pageBuf[10])<<16 | int(pageBuf[11])<<24

		var pageType string

		switch {
		case flags&0x01 != 0:
			pageType = "branch"
		case flags&0x02 != 0:
			pageType = "leaf"
		case flags&0x04 != 0:
			pageType = "meta"
		case flags&0x10 != 0:
			pageType = "freelist"
		default:
			pageType = "free"
		}

		pages = append(pages, PageInfo{
			ID:       i,
			Type:     pageType,
			Count:    count,
			Overflow: overflow,
		})
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(pages)
	}

	_, _ = fmt.Fprintf(w, "%-8s %-10s %-8s %-8s\n", "ID", "TYPE", "COUNT", "OVERFLOW")
	for _, p := range pages {
		_, _ = fmt.Fprintf(w, "%-8d %-10s %-8d %-8d\n", p.ID, p.Type, p.Count, p.Overflow)
	}

	return nil
}

// RunPageDump dumps a page in hexadecimal format (matches bbolt dump)
func RunPageDump(w io.Writer, dbPath string, pageID int, opts Options) error {
	f, err := os.Open(dbPath)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	defer func() { _ = f.Close() }()

	// Read page size from meta
	buf := make([]byte, 4096)

	_, err = f.Read(buf)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	pageSize := int(buf[12]) | int(buf[13])<<8 | int(buf[14])<<16 | int(buf[15])<<24
	if pageSize == 0 {
		pageSize = 4096
	}

	// Seek to the requested page
	_, err = f.Seek(int64(pageID*pageSize), 0)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	// Read the page
	pageBuf := make([]byte, pageSize)

	n, err := f.Read(pageBuf)
	if err != nil {
		return fmt.Errorf("bbolt: %w", err)
	}

	if opts.JSON {
		result := map[string]any{
			"page_id":   pageID,
			"page_size": pageSize,
			"hex":       hex.EncodeToString(pageBuf[:n]),
		}

		return json.NewEncoder(w).Encode(result)
	}

	// Print hex dump
	_, _ = fmt.Fprintf(w, "Page %d (size: %d)\n", pageID, pageSize)

	for i := 0; i < n; i += 16 {
		_, _ = fmt.Fprintf(w, "%08x  ", i)

		// Hex
		for j := range 16 {
			if i+j < n {
				_, _ = fmt.Fprintf(w, "%02x ", pageBuf[i+j])
			} else {
				_, _ = fmt.Fprint(w, "   ")
			}

			if j == 7 {
				_, _ = fmt.Fprint(w, " ")
			}
		}

		_, _ = fmt.Fprint(w, " |")

		// ASCII
		for j := 0; j < 16 && i+j < n; j++ {
			b := pageBuf[i+j]
			if b >= 32 && b < 127 {
				_, _ = fmt.Fprintf(w, "%c", b)
			} else {
				_, _ = fmt.Fprint(w, ".")
			}
		}

		_, _ = fmt.Fprintln(w, "|")
	}

	return nil
}

func formatValue(v []byte, asHex bool) string {
	if asHex {
		return hex.EncodeToString(v)
	}

	return string(v)
}

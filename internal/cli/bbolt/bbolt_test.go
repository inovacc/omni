package bbolt

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func createTestDB(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "bbolt_test")
	if err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := bolt.Open(dbPath, 0600, nil)

	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	// Create test buckets and data
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("users"))
		if err != nil {
			return err
		}

		_ = b.Put([]byte("user1"), []byte("Alice"))
		_ = b.Put([]byte("user2"), []byte("Bob"))
		_ = b.Put([]byte("user3"), []byte("Charlie"))

		cfg, err := tx.CreateBucket([]byte("config"))
		if err != nil {
			return err
		}

		_ = cfg.Put([]byte("version"), []byte("1.0.0"))
		_ = cfg.Put([]byte("debug"), []byte("true"))

		return nil
	})

	if err != nil {
		_ = db.Close()
		_ = os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	_ = db.Close()

	return dbPath, func() { _ = os.RemoveAll(tmpDir) }
}

func TestRunStats(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("text output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunStats(&buf, dbPath, Options{})
		if err != nil {
			t.Fatalf("RunStats() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Path:") {
			t.Error("RunStats() should contain Path")
		}

		if !strings.Contains(output, "Size:") {
			t.Error("RunStats() should contain Size")
		}

		if !strings.Contains(output, "Buckets:") {
			t.Error("RunStats() should contain Buckets")
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunStats(&buf, dbPath, Options{JSON: true})
		if err != nil {
			t.Fatalf("RunStats() error = %v", err)
		}

		var result StatsResult
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("RunStats() invalid JSON: %v", err)
		}

		if result.Buckets != 2 {
			t.Errorf("RunStats() buckets = %d, want 2", result.Buckets)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunStats(&buf, "/nonexistent/db.bolt", Options{})
		if err == nil {
			t.Error("RunStats() should error for nonexistent file")
		}
	})
}

func TestRunBuckets(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("text output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunBuckets(&buf, dbPath, Options{})
		if err != nil {
			t.Fatalf("RunBuckets() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "users") {
			t.Error("RunBuckets() should list users bucket")
		}

		if !strings.Contains(output, "config") {
			t.Error("RunBuckets() should list config bucket")
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunBuckets(&buf, dbPath, Options{JSON: true})
		if err != nil {
			t.Fatalf("RunBuckets() error = %v", err)
		}

		var buckets []BucketInfo
		if err := json.Unmarshal(buf.Bytes(), &buckets); err != nil {
			t.Fatalf("RunBuckets() invalid JSON: %v", err)
		}

		if len(buckets) != 2 {
			t.Errorf("RunBuckets() got %d buckets, want 2", len(buckets))
		}
	})
}

func TestRunKeys(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("list all keys", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKeys(&buf, dbPath, "users", Options{})
		if err != nil {
			t.Fatalf("RunKeys() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "user1") {
			t.Error("RunKeys() should list user1")
		}

		if !strings.Contains(output, "user2") {
			t.Error("RunKeys() should list user2")
		}
	})

	t.Run("with prefix filter", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKeys(&buf, dbPath, "users", Options{Prefix: "user1"})
		if err != nil {
			t.Fatalf("RunKeys() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 {
			t.Errorf("RunKeys() with prefix should return 1 key, got %d", len(lines))
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKeys(&buf, dbPath, "users", Options{JSON: true})
		if err != nil {
			t.Fatalf("RunKeys() error = %v", err)
		}

		var keys []string
		if err := json.Unmarshal(buf.Bytes(), &keys); err != nil {
			t.Fatalf("RunKeys() invalid JSON: %v", err)
		}

		if len(keys) != 3 {
			t.Errorf("RunKeys() got %d keys, want 3", len(keys))
		}
	})

	t.Run("nonexistent bucket", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKeys(&buf, dbPath, "nonexistent", Options{})
		if err == nil {
			t.Error("RunKeys() should error for nonexistent bucket")
		}
	})
}

func TestRunGet(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("get existing key", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunGet(&buf, dbPath, "users", "user1", Options{})
		if err != nil {
			t.Fatalf("RunGet() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "Alice" {
			t.Errorf("RunGet() = %v, want Alice", output)
		}
	})

	t.Run("get with hex", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunGet(&buf, dbPath, "users", "user1", Options{Hex: true})
		if err != nil {
			t.Fatalf("RunGet() error = %v", err)
		}

		// "Alice" in hex is "416c696365"
		output := strings.TrimSpace(buf.String())
		if output != "416c696365" {
			t.Errorf("RunGet() hex = %v, want 416c696365", output)
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunGet(&buf, dbPath, "users", "user1", Options{JSON: true})
		if err != nil {
			t.Fatalf("RunGet() error = %v", err)
		}

		var kv KeyValue
		if err := json.Unmarshal(buf.Bytes(), &kv); err != nil {
			t.Fatalf("RunGet() invalid JSON: %v", err)
		}

		if kv.Key != "user1" || kv.Value != "Alice" {
			t.Errorf("RunGet() = %+v, want user1/Alice", kv)
		}
	})

	t.Run("nonexistent key", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunGet(&buf, dbPath, "users", "nonexistent", Options{})
		if err == nil {
			t.Error("RunGet() should error for nonexistent key")
		}
	})
}

func TestRunPut(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("put new key", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunPut(&buf, dbPath, "users", "user4", "David", Options{})
		if err != nil {
			t.Fatalf("RunPut() error = %v", err)
		}

		// Verify key was set
		var getBuf bytes.Buffer
		err = RunGet(&getBuf, dbPath, "users", "user4", Options{})
		if err != nil {
			t.Fatalf("RunGet() after put error = %v", err)
		}

		if strings.TrimSpace(getBuf.String()) != "David" {
			t.Errorf("RunPut() did not store value correctly")
		}
	})

	t.Run("put to new bucket", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunPut(&buf, dbPath, "newbucket", "key1", "value1", Options{})
		if err != nil {
			t.Fatalf("RunPut() error = %v", err)
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunPut(&buf, dbPath, "users", "user5", "Eve", Options{JSON: true})
		if err != nil {
			t.Fatalf("RunPut() error = %v", err)
		}

		var result map[string]string
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("RunPut() invalid JSON: %v", err)
		}

		if result["status"] != "ok" {
			t.Errorf("RunPut() status = %v, want ok", result["status"])
		}
	})
}

func TestRunDelete(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("delete existing key", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDelete(&buf, dbPath, "users", "user1", Options{})
		if err != nil {
			t.Fatalf("RunDelete() error = %v", err)
		}

		// Verify key was deleted
		var getBuf bytes.Buffer
		err = RunGet(&getBuf, dbPath, "users", "user1", Options{})
		if err == nil {
			t.Error("RunDelete() key should be deleted")
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDelete(&buf, dbPath, "users", "user2", Options{JSON: true})
		if err != nil {
			t.Fatalf("RunDelete() error = %v", err)
		}

		var result map[string]string
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("RunDelete() invalid JSON: %v", err)
		}

		if result["status"] != "deleted" {
			t.Errorf("RunDelete() status = %v, want deleted", result["status"])
		}
	})
}

func TestRunDump(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("dump bucket", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDump(&buf, dbPath, "users", Options{})
		if err != nil {
			t.Fatalf("RunDump() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "user1") {
			t.Error("RunDump() should contain user1")
		}

		if !strings.Contains(output, "Alice") {
			t.Error("RunDump() should contain Alice")
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDump(&buf, dbPath, "users", Options{JSON: true})
		if err != nil {
			t.Fatalf("RunDump() error = %v", err)
		}

		var kvPairs []KeyValue
		if err := json.Unmarshal(buf.Bytes(), &kvPairs); err != nil {
			t.Fatalf("RunDump() invalid JSON: %v", err)
		}

		if len(kvPairs) != 3 {
			t.Errorf("RunDump() got %d pairs, want 3", len(kvPairs))
		}
	})
}

func TestRunCompact(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("compact database", func(t *testing.T) {
		dstPath := dbPath + ".compact"
		defer func() { _ = os.Remove(dstPath) }()

		var buf bytes.Buffer

		err := RunCompact(&buf, dbPath, dstPath, Options{})
		if err != nil {
			t.Fatalf("RunCompact() error = %v", err)
		}

		// Verify compacted database exists
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Error("RunCompact() should create destination file")
		}
	})

	t.Run("json output", func(t *testing.T) {
		dstPath := dbPath + ".compact2"
		defer func() { _ = os.Remove(dstPath) }()

		var buf bytes.Buffer

		err := RunCompact(&buf, dbPath, dstPath, Options{JSON: true})
		if err != nil {
			t.Fatalf("RunCompact() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("RunCompact() invalid JSON: %v", err)
		}

		if result["status"] != "ok" {
			t.Errorf("RunCompact() status = %v, want ok", result["status"])
		}
	})
}

func TestRunCheck(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("check valid database", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunCheck(&buf, dbPath, Options{})
		if err != nil {
			t.Fatalf("RunCheck() error = %v", err)
		}

		if !strings.Contains(buf.String(), "OK") {
			t.Error("RunCheck() should report OK for valid database")
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunCheck(&buf, dbPath, Options{JSON: true})
		if err != nil {
			t.Fatalf("RunCheck() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("RunCheck() invalid JSON: %v", err)
		}

		if result["status"] != "ok" {
			t.Errorf("RunCheck() status = %v, want ok", result["status"])
		}
	})
}

func TestRunCreateBucket(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("create new bucket", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunCreateBucket(&buf, dbPath, "newbucket", Options{})
		if err != nil {
			t.Fatalf("RunCreateBucket() error = %v", err)
		}

		// Verify bucket was created
		var listBuf bytes.Buffer
		_ = RunBuckets(&listBuf, dbPath, Options{})

		if !strings.Contains(listBuf.String(), "newbucket") {
			t.Error("RunCreateBucket() should create bucket")
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunCreateBucket(&buf, dbPath, "jsonbucket", Options{JSON: true})
		if err != nil {
			t.Fatalf("RunCreateBucket() error = %v", err)
		}

		var result map[string]string
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("RunCreateBucket() invalid JSON: %v", err)
		}

		if result["status"] != "created" {
			t.Errorf("RunCreateBucket() status = %v, want created", result["status"])
		}
	})
}

func TestRunDeleteBucket(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("delete existing bucket", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDeleteBucket(&buf, dbPath, "config", Options{})
		if err != nil {
			t.Fatalf("RunDeleteBucket() error = %v", err)
		}

		// Verify bucket was deleted
		var listBuf bytes.Buffer
		_ = RunBuckets(&listBuf, dbPath, Options{})

		if strings.Contains(listBuf.String(), "config") {
			t.Error("RunDeleteBucket() should delete bucket")
		}
	})

	t.Run("delete nonexistent bucket", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDeleteBucket(&buf, dbPath, "nonexistent", Options{})
		if err == nil {
			t.Error("RunDeleteBucket() should error for nonexistent bucket")
		}
	})
}

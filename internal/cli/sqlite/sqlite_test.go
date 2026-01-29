package sqlite

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func createTestDB(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "sqlite_test")
	if err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sql.Open("sqlite", dbPath)

	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	// Create test tables and data
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE
		);

		INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com');
		INSERT INTO users (name, email) VALUES ('Bob', 'bob@example.com');
		INSERT INTO users (name, email) VALUES ('Charlie', 'charlie@example.com');

		CREATE TABLE config (
			key TEXT PRIMARY KEY,
			value TEXT
		);

		INSERT INTO config VALUES ('version', '1.0.0');
		INSERT INTO config VALUES ('debug', 'true');

		CREATE INDEX idx_users_email ON users(email);

		CREATE VIEW active_users AS SELECT * FROM users WHERE id > 0;
	`)

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

		if !strings.Contains(output, "Tables:") {
			t.Error("RunStats() should contain Tables")
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

		if result.Tables != 2 {
			t.Errorf("RunStats() tables = %d, want 2", result.Tables)
		}
	})
}

func TestRunTables(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("text output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunTables(&buf, dbPath, Options{})
		if err != nil {
			t.Fatalf("RunTables() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "users") {
			t.Error("RunTables() should list users table")
		}

		if !strings.Contains(output, "config") {
			t.Error("RunTables() should list config table")
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunTables(&buf, dbPath, Options{JSON: true})
		if err != nil {
			t.Fatalf("RunTables() error = %v", err)
		}

		var tables []TableInfo
		if err := json.Unmarshal(buf.Bytes(), &tables); err != nil {
			t.Fatalf("RunTables() invalid JSON: %v", err)
		}

		// 2 tables + 1 view
		if len(tables) != 3 {
			t.Errorf("RunTables() got %d tables, want 3", len(tables))
		}
	})
}

func TestRunSchema(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("show all schemas", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSchema(&buf, dbPath, "", Options{})
		if err != nil {
			t.Fatalf("RunSchema() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "CREATE TABLE") {
			t.Error("RunSchema() should contain CREATE TABLE")
		}
	})

	t.Run("show specific table", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSchema(&buf, dbPath, "users", Options{})
		if err != nil {
			t.Fatalf("RunSchema() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "users") {
			t.Error("RunSchema() should contain table name")
		}
	})

	t.Run("nonexistent table", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSchema(&buf, dbPath, "nonexistent", Options{})
		if err == nil {
			t.Error("RunSchema() should error for nonexistent table")
		}
	})
}

func TestRunColumns(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("show columns", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunColumns(&buf, dbPath, "users", Options{})
		if err != nil {
			t.Fatalf("RunColumns() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "id") {
			t.Error("RunColumns() should contain id column")
		}

		if !strings.Contains(output, "name") {
			t.Error("RunColumns() should contain name column")
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunColumns(&buf, dbPath, "users", Options{JSON: true})
		if err != nil {
			t.Fatalf("RunColumns() error = %v", err)
		}

		var columns []ColumnInfo
		if err := json.Unmarshal(buf.Bytes(), &columns); err != nil {
			t.Fatalf("RunColumns() invalid JSON: %v", err)
		}

		if len(columns) != 3 {
			t.Errorf("RunColumns() got %d columns, want 3", len(columns))
		}
	})
}

func TestRunIndexes(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("list indexes", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunIndexes(&buf, dbPath, Options{})
		if err != nil {
			t.Fatalf("RunIndexes() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "idx_users_email") {
			t.Error("RunIndexes() should list idx_users_email")
		}
	})
}

func TestRunQuery(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("select query", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunQuery(&buf, dbPath, "SELECT name FROM users WHERE id = 1", Options{})
		if err != nil {
			t.Fatalf("RunQuery() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Alice") {
			t.Error("RunQuery() should return Alice")
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunQuery(&buf, dbPath, "SELECT * FROM users ORDER BY id", Options{JSON: true})
		if err != nil {
			t.Fatalf("RunQuery() error = %v", err)
		}

		var results []map[string]any
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Fatalf("RunQuery() invalid JSON: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("RunQuery() got %d rows, want 3", len(results))
		}
	})

	t.Run("insert query", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunQuery(&buf, dbPath, "INSERT INTO users (name, email) VALUES ('Dave', 'dave@example.com')", Options{})
		if err != nil {
			t.Fatalf("RunQuery() error = %v", err)
		}

		if !strings.Contains(buf.String(), "1") {
			t.Error("RunQuery() should show 1 row affected")
		}
	})

	t.Run("with header", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunQuery(&buf, dbPath, "SELECT name, email FROM users LIMIT 1", Options{Header: true})
		if err != nil {
			t.Fatalf("RunQuery() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "name") || !strings.Contains(output, "email") {
			t.Error("RunQuery() with header should show column names")
		}
	})
}

func TestRunVacuum(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("vacuum database", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunVacuum(&buf, dbPath, Options{})
		if err != nil {
			t.Fatalf("RunVacuum() error = %v", err)
		}

		if !strings.Contains(buf.String(), "complete") {
			t.Error("RunVacuum() should report complete")
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunVacuum(&buf, dbPath, Options{JSON: true})
		if err != nil {
			t.Fatalf("RunVacuum() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("RunVacuum() invalid JSON: %v", err)
		}

		if result["status"] != "ok" {
			t.Errorf("RunVacuum() status = %v, want ok", result["status"])
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
			t.Error("RunCheck() should report OK")
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

func TestRunDump(t *testing.T) {
	dbPath, cleanup := createTestDB(t)
	defer cleanup()

	t.Run("dump all tables", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDump(&buf, dbPath, "", Options{})
		if err != nil {
			t.Fatalf("RunDump() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "BEGIN TRANSACTION") {
			t.Error("RunDump() should contain BEGIN TRANSACTION")
		}

		if !strings.Contains(output, "CREATE TABLE") {
			t.Error("RunDump() should contain CREATE TABLE")
		}

		if !strings.Contains(output, "INSERT INTO") {
			t.Error("RunDump() should contain INSERT INTO")
		}

		if !strings.Contains(output, "COMMIT") {
			t.Error("RunDump() should contain COMMIT")
		}
	})

	t.Run("dump specific table", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDump(&buf, dbPath, "users", Options{})
		if err != nil {
			t.Fatalf("RunDump() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "users") {
			t.Error("RunDump() should contain users table")
		}

		if strings.Contains(output, "config") {
			t.Error("RunDump() should not contain config table")
		}
	})
}

func TestRunImport(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sqlite_import_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	dbPath := filepath.Join(tmpDir, "import.db")
	sqlFile := filepath.Join(tmpDir, "import.sql")

	// Create SQL file
	sqlContent := `CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT);
INSERT INTO test VALUES (1, 'test');`

	if err := os.WriteFile(sqlFile, []byte(sqlContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("import sql file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunImport(&buf, dbPath, sqlFile, Options{})
		if err != nil {
			t.Fatalf("RunImport() error = %v", err)
		}

		// Verify data was imported
		var verifyBuf bytes.Buffer
		err = RunQuery(&verifyBuf, dbPath, "SELECT name FROM test", Options{})

		if err != nil {
			t.Fatalf("Verify query error = %v", err)
		}

		if !strings.Contains(verifyBuf.String(), "test") {
			t.Error("RunImport() should import data")
		}
	})
}

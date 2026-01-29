package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// Options configures sqlite command behavior
type Options struct {
	JSON      bool   // --json: output as JSON
	Header    bool   // --header: show column headers
	Separator string // --separator: column separator
	Mode      string // --mode: output mode (column, csv, line, etc)
}

// StatsResult represents database statistics for JSON output
type StatsResult struct {
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	Tables     int    `json:"tables"`
	Indexes    int    `json:"indexes"`
	Views      int    `json:"views"`
	Triggers   int    `json:"triggers"`
	PageSize   int64  `json:"page_size"`
	PageCount  int64  `json:"page_count"`
	FreePages  int64  `json:"free_pages"`
	SchemaVer  int64  `json:"schema_version"`
	UserVer    int64  `json:"user_version"`
	WALEnabled bool   `json:"wal_enabled"`
}

// TableInfo represents table information for JSON output
type TableInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	RowCount int64  `json:"row_count,omitempty"`
}

// ColumnInfo represents column information for JSON output
type ColumnInfo struct {
	CID        int    `json:"cid"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	NotNull    bool   `json:"notnull"`
	Default    any    `json:"default"`
	PrimaryKey bool   `json:"pk"`
}

// IndexInfo represents index information for JSON output
type IndexInfo struct {
	Name    string `json:"name"`
	Table   string `json:"table"`
	Unique  bool   `json:"unique"`
	Columns string `json:"columns"`
}

// RunStats displays database statistics
func RunStats(w io.Writer, dbPath string, opts Options) error {
	fi, err := os.Stat(dbPath)
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = db.Close() }()

	result := StatsResult{
		Path: filepath.Clean(dbPath),
		Size: fi.Size(),
	}

	// Count tables, views, indexes, triggers
	var count int64
	_ = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&count)
	result.Tables = int(count)

	_ = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='index'").Scan(&count)
	result.Indexes = int(count)

	_ = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='view'").Scan(&count)
	result.Views = int(count)

	_ = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='trigger'").Scan(&count)
	result.Triggers = int(count)

	// Page info
	_ = db.QueryRow("PRAGMA page_size").Scan(&result.PageSize)
	_ = db.QueryRow("PRAGMA page_count").Scan(&result.PageCount)
	_ = db.QueryRow("PRAGMA freelist_count").Scan(&result.FreePages)
	_ = db.QueryRow("PRAGMA schema_version").Scan(&result.SchemaVer)
	_ = db.QueryRow("PRAGMA user_version").Scan(&result.UserVer)

	var journalMode string
	_ = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	result.WALEnabled = journalMode == "wal"

	if opts.JSON {
		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Path: %s\n", result.Path)
	_, _ = fmt.Fprintf(w, "Size: %d bytes\n", result.Size)
	_, _ = fmt.Fprintf(w, "Tables: %d\n", result.Tables)
	_, _ = fmt.Fprintf(w, "Indexes: %d\n", result.Indexes)
	_, _ = fmt.Fprintf(w, "Views: %d\n", result.Views)
	_, _ = fmt.Fprintf(w, "Triggers: %d\n", result.Triggers)
	_, _ = fmt.Fprintf(w, "Page Size: %d\n", result.PageSize)
	_, _ = fmt.Fprintf(w, "Page Count: %d\n", result.PageCount)
	_, _ = fmt.Fprintf(w, "Free Pages: %d\n", result.FreePages)
	_, _ = fmt.Fprintf(w, "Schema Version: %d\n", result.SchemaVer)
	_, _ = fmt.Fprintf(w, "User Version: %d\n", result.UserVer)
	_, _ = fmt.Fprintf(w, "WAL Enabled: %v\n", result.WALEnabled)

	return nil
}

// RunTables lists all tables in the database
func RunTables(w io.Writer, dbPath string, opts Options) error {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = db.Close() }()

	rows, err := db.Query("SELECT name, type FROM sqlite_master WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var tables []TableInfo

	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.Type); err != nil {
			continue
		}

		// Get row count for tables
		if t.Type == "table" {
			_ = db.QueryRow(fmt.Sprintf("SELECT count(*) FROM %q", t.Name)).Scan(&t.RowCount)
		}

		tables = append(tables, t)
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(tables)
	}

	for _, t := range tables {
		if t.Type == "view" {
			_, _ = fmt.Fprintf(w, "%s (view)\n", t.Name)
		} else {
			_, _ = fmt.Fprintf(w, "%s (%d rows)\n", t.Name, t.RowCount)
		}
	}

	return nil
}

// RunSchema shows table schema
func RunSchema(w io.Writer, dbPath, table string, opts Options) error {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = db.Close() }()

	if table == "" {
		// Show all schemas
		rows, err := db.Query("SELECT sql FROM sqlite_master WHERE sql IS NOT NULL ORDER BY type, name")
		if err != nil {
			return fmt.Errorf("sqlite: %w", err)
		}

		defer func() { _ = rows.Close() }()

		var schemas []string

		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				continue
			}

			schemas = append(schemas, s)
		}

		if opts.JSON {
			return json.NewEncoder(w).Encode(schemas)
		}

		for _, s := range schemas {
			_, _ = fmt.Fprintln(w, s+";")
		}

		return nil
	}

	// Show specific table schema
	var sqlStr string
	err = db.QueryRow("SELECT sql FROM sqlite_master WHERE name = ? AND sql IS NOT NULL", table).Scan(&sqlStr)

	if err != nil {
		return fmt.Errorf("sqlite: table %q not found", table)
	}

	if opts.JSON {
		result := map[string]string{
			"table":  table,
			"schema": sqlStr,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintln(w, sqlStr+";")

	return nil
}

// RunColumns shows table columns
func RunColumns(w io.Writer, dbPath, table string, opts Options) error {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = db.Close() }()

	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%q)", table))
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var columns []ColumnInfo

	for rows.Next() {
		var c ColumnInfo
		var notNull, pk int
		var dflt sql.NullString

		if err := rows.Scan(&c.CID, &c.Name, &c.Type, &notNull, &dflt, &pk); err != nil {
			continue
		}

		c.NotNull = notNull != 0
		c.PrimaryKey = pk != 0

		if dflt.Valid {
			c.Default = dflt.String
		}

		columns = append(columns, c)
	}

	if len(columns) == 0 {
		return fmt.Errorf("sqlite: table %q not found", table)
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(columns)
	}

	_, _ = fmt.Fprintf(w, "%-4s %-20s %-15s %-8s %-8s %-10s\n", "CID", "NAME", "TYPE", "NOTNULL", "PK", "DEFAULT")

	for _, c := range columns {
		dflt := ""
		if c.Default != nil {
			dflt = fmt.Sprintf("%v", c.Default)
		}

		_, _ = fmt.Fprintf(w, "%-4d %-20s %-15s %-8v %-8v %-10s\n", c.CID, c.Name, c.Type, c.NotNull, c.PrimaryKey, dflt)
	}

	return nil
}

// RunIndexes lists all indexes
func RunIndexes(w io.Writer, dbPath string, opts Options) error {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = db.Close() }()

	rows, err := db.Query(`SELECT name, tbl_name, sql FROM sqlite_master WHERE type='index' AND sql IS NOT NULL ORDER BY tbl_name, name`)
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var indexes []IndexInfo

	for rows.Next() {
		var idx IndexInfo
		var sqlStr string

		if err := rows.Scan(&idx.Name, &idx.Table, &sqlStr); err != nil {
			continue
		}

		idx.Unique = strings.Contains(strings.ToUpper(sqlStr), "UNIQUE")

		// Extract column names from SQL
		start := strings.LastIndex(sqlStr, "(")
		end := strings.LastIndex(sqlStr, ")")

		if start >= 0 && end > start {
			idx.Columns = sqlStr[start+1 : end]
		}

		indexes = append(indexes, idx)
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(indexes)
	}

	_, _ = fmt.Fprintf(w, "%-30s %-20s %-8s %-30s\n", "INDEX", "TABLE", "UNIQUE", "COLUMNS")

	for _, idx := range indexes {
		_, _ = fmt.Fprintf(w, "%-30s %-20s %-8v %-30s\n", idx.Name, idx.Table, idx.Unique, idx.Columns)
	}

	return nil
}

// RunQuery executes a SQL query and displays results
func RunQuery(w io.Writer, dbPath, query string, opts Options) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = db.Close() }()

	// Check if it's a SELECT query
	queryUpper := strings.TrimSpace(strings.ToUpper(query))
	isSelect := strings.HasPrefix(queryUpper, "SELECT") || strings.HasPrefix(queryUpper, "PRAGMA") || strings.HasPrefix(queryUpper, "EXPLAIN")

	if !isSelect {
		// Execute non-SELECT query
		result, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("sqlite: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()

		if opts.JSON {
			return json.NewEncoder(w).Encode(map[string]int64{"rows_affected": rowsAffected})
		}

		_, _ = fmt.Fprintf(w, "Rows affected: %d\n", rowsAffected)

		return nil
	}

	// Execute SELECT query
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	var results []map[string]any

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = values[i]
		}

		results = append(results, row)
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(results)
	}

	separator := opts.Separator
	if separator == "" {
		separator = "|"
	}

	// Print header
	if opts.Header {
		_, _ = fmt.Fprintln(w, strings.Join(columns, separator))
	}

	// Print rows
	for _, row := range results {
		var vals []string
		for _, col := range columns {
			vals = append(vals, fmt.Sprintf("%v", row[col]))
		}

		_, _ = fmt.Fprintln(w, strings.Join(vals, separator))
	}

	return nil
}

// RunVacuum optimizes the database
func RunVacuum(w io.Writer, dbPath string, opts Options) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = db.Close() }()

	// Get size before
	fi, _ := os.Stat(dbPath)
	sizeBefore := fi.Size()

	_, err = db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	// Get size after
	fi, _ = os.Stat(dbPath)
	sizeAfter := fi.Size()

	if opts.JSON {
		result := map[string]any{
			"status":      "ok",
			"size_before": sizeBefore,
			"size_after":  sizeAfter,
			"savings":     sizeBefore - sizeAfter,
		}

		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Vacuum complete\n")
	_, _ = fmt.Fprintf(w, "Before: %d bytes\n", sizeBefore)
	_, _ = fmt.Fprintf(w, "After: %d bytes\n", sizeAfter)
	_, _ = fmt.Fprintf(w, "Savings: %d bytes\n", sizeBefore-sizeAfter)

	return nil
}

// RunCheck verifies database integrity
func RunCheck(w io.Writer, dbPath string, opts Options) error {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = db.Close() }()

	rows, err := db.Query("PRAGMA integrity_check")
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var results []string

	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			continue
		}

		results = append(results, s)
	}

	isOK := len(results) == 1 && results[0] == "ok"

	if opts.JSON {
		result := map[string]any{
			"status": "ok",
			"errors": []string{},
		}

		if !isOK {
			result["status"] = "errors_found"
			result["errors"] = results
		}

		return json.NewEncoder(w).Encode(result)
	}

	if isOK {
		_, _ = fmt.Fprintln(w, "Database OK")
	} else {
		_, _ = fmt.Fprintln(w, "Errors found:")
		for _, r := range results {
			_, _ = fmt.Fprintf(w, "  %s\n", r)
		}
	}

	return nil
}

// RunDump exports database as SQL
func RunDump(w io.Writer, dbPath, table string, opts Options) error {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = db.Close() }()

	var tables []string

	if table != "" {
		tables = []string{table}
	} else {
		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
		if err != nil {
			return fmt.Errorf("sqlite: %w", err)
		}

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				continue
			}

			tables = append(tables, name)
		}

		_ = rows.Close()
	}

	_, _ = fmt.Fprintln(w, "BEGIN TRANSACTION;")

	for _, t := range tables {
		// Get CREATE statement
		var sqlStr string
		err := db.QueryRow("SELECT sql FROM sqlite_master WHERE name = ? AND sql IS NOT NULL", t).Scan(&sqlStr)

		if err == nil {
			_, _ = fmt.Fprintf(w, "%s;\n", sqlStr)
		}

		// Get data
		rows, err := db.Query(fmt.Sprintf("SELECT * FROM %q", t))
		if err != nil {
			continue
		}

		columns, _ := rows.Columns()

		for rows.Next() {
			values := make([]any, len(columns))
			valuePtrs := make([]any, len(columns))

			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				continue
			}

			var vals []string

			for _, v := range values {
				switch val := v.(type) {
				case nil:
					vals = append(vals, "NULL")
				case []byte:
					vals = append(vals, fmt.Sprintf("'%s'", strings.ReplaceAll(string(val), "'", "''")))
				case string:
					vals = append(vals, fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "''")))
				default:
					vals = append(vals, fmt.Sprintf("%v", val))
				}
			}

			_, _ = fmt.Fprintf(w, "INSERT INTO %q VALUES(%s);\n", t, strings.Join(vals, ","))
		}

		_ = rows.Close()
	}

	_, _ = fmt.Fprintln(w, "COMMIT;")

	return nil
}

// RunImport imports SQL file into database
func RunImport(w io.Writer, dbPath, sqlFile string, opts Options) error {
	content, err := os.ReadFile(sqlFile)
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	defer func() { _ = db.Close() }()

	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(map[string]string{"status": "ok", "file": sqlFile})
	}

	_, _ = fmt.Fprintf(w, "Imported %s\n", sqlFile)

	return nil
}

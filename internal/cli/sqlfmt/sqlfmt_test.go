package sqlfmt

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    Options
		want    string
		wantErr bool
	}{
		{
			name:  "simple select",
			input: "select * from users",
			opts:  Options{Uppercase: true},
			want:  "SELECT *\nFROM users",
		},
		{
			name:  "select with where",
			input: "select id, name from users where active = true",
			opts:  Options{Uppercase: true},
			want:  "SELECT id,\nname\nFROM users\nWHERE active = TRUE",
		},
		{
			name:  "select with join",
			input: "select u.id, o.amount from users u join orders o on u.id = o.user_id",
			opts:  Options{Uppercase: true},
			want:  "SELECT u.id,\no.amount\nFROM users u\nJOIN orders o\nON u.id = o.user_id",
		},
		{
			name:  "insert statement",
			input: "insert into users (name, email) values ('John', 'john@example.com')",
			opts:  Options{Uppercase: true},
			want:  "INSERT INTO users (name,\n  email)\nVALUES ('John',\n  'john@example.com')",
		},
		{
			name:  "update statement",
			input: "update users set name = 'Jane' where id = 1",
			opts:  Options{Uppercase: true},
			want:  "UPDATE users\nSET name = 'Jane'\nWHERE id = 1",
		},
		{
			name:  "delete statement",
			input: "delete from users where id = 1",
			opts:  Options{Uppercase: true},
			want:  "DELETE\nFROM users\nWHERE id = 1",
		},
		{
			name:  "preserve lowercase",
			input: "select * from users",
			opts:  Options{Uppercase: false},
			want:  "select *\nfrom users",
		},
		{
			name:  "with order by",
			input: "select * from users order by name asc",
			opts:  Options{Uppercase: true},
			want:  "SELECT *\nFROM users\nORDER BY name ASC",
		},
		{
			name:  "with group by",
			input: "select count(*), status from orders group by status",
			opts:  Options{Uppercase: true},
			want:  "SELECT COUNT(*),\nstatus\nFROM orders\nGROUP BY status",
		},
		{
			name:  "with having",
			input: "select status, count(*) from orders group by status having count(*) > 10",
			opts:  Options{Uppercase: true},
			want:  "SELECT status,\nCOUNT(*)\nFROM orders\nGROUP BY status\nHAVING COUNT(*) > 10",
		},
		{
			name:  "with limit",
			input: "select * from users limit 10 offset 20",
			opts:  Options{Uppercase: true},
			want:  "SELECT *\nFROM users\nLIMIT 10\nOFFSET 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := Run(&buf, r, nil, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := strings.TrimSpace(buf.String())
			if got != tt.want {
				t.Errorf("Run() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestRunMinify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple select",
			input: "SELECT * \nFROM users \nWHERE id = 1",
			want:  "SELECT * FROM users WHERE id = 1",
		},
		{
			name:  "with extra whitespace",
			input: "SELECT    *     FROM    users    WHERE    id   =   1",
			want:  "SELECT * FROM users WHERE id = 1",
		},
		{
			name:  "multiline",
			input: "SELECT\n    id,\n    name\nFROM\n    users",
			want:  "SELECT id,name FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := RunMinify(&buf, r, nil, Options{})
			if err != nil {
				t.Errorf("RunMinify() error = %v", err)
				return
			}

			got := strings.TrimSpace(buf.String())
			if got != tt.want {
				t.Errorf("RunMinify() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid select",
			input:   "SELECT * FROM users",
			wantErr: false,
		},
		{
			name:    "valid insert",
			input:   "INSERT INTO users (name) VALUES ('John')",
			wantErr: false,
		},
		{
			name:    "valid update",
			input:   "UPDATE users SET name = 'Jane' WHERE id = 1",
			wantErr: false,
		},
		{
			name:    "valid delete",
			input:   "DELETE FROM users WHERE id = 1",
			wantErr: false,
		},
		{
			name:    "valid create",
			input:   "CREATE TABLE users (id INT PRIMARY KEY)",
			wantErr: false,
		},
		{
			name:    "unbalanced parens open",
			input:   "SELECT * FROM users WHERE (id = 1",
			wantErr: true,
		},
		{
			name:    "unbalanced parens close",
			input:   "SELECT * FROM users WHERE id = 1)",
			wantErr: true,
		},
		{
			name:    "unbalanced quotes",
			input:   "SELECT * FROM users WHERE name = 'John",
			wantErr: true,
		},
		{
			name:    "invalid start",
			input:   "INVALID FROM users",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "comment only",
			input:   "-- this is a comment",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := RunValidate(&buf, r, nil, ValidateOptions{})
			if (err != nil) != tt.wantErr {
				t.Errorf("RunValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsKeyword(t *testing.T) {
	tests := []struct {
		token string
		want  bool
	}{
		{"SELECT", true},
		{"select", true},
		{"FROM", true},
		{"WHERE", true},
		{"users", false},
		{"id", false},
		{"123", false},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			got := isKeyword(tt.token)
			if got != tt.want {
				t.Errorf("isKeyword(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}

func TestNeedsSpace(t *testing.T) {
	tests := []struct {
		prev string
		curr string
		want bool
	}{
		{"SELECT", "*", true},
		{"(", "id", false},
		{"id", ")", false},
		{"id", ",", false},
		{",", "name", false},
		{"table", ".", false},
		{".", "column", false},
		{"", "SELECT", false},
		{"SELECT", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.prev+"_"+tt.curr, func(t *testing.T) {
			got := needsSpace(tt.prev, tt.curr)
			if got != tt.want {
				t.Errorf("needsSpace(%q, %q) = %v, want %v", tt.prev, tt.curr, got, tt.want)
			}
		})
	}
}

func TestTokenizeSQL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple select",
			input: "SELECT * FROM users",
			want:  []string{"SELECT", "*", "FROM", "users"},
		},
		{
			name:  "with operators",
			input: "id = 1",
			want:  []string{"id", "=", "1"},
		},
		{
			name:  "with string literal",
			input: "name = 'John'",
			want:  []string{"name", "=", "'John'"},
		},
		{
			name:  "with parens",
			input: "(id, name)",
			want:  []string{"(", "id", ",", "name", ")"},
		},
		{
			name:  "comparison operators",
			input: "a >= b AND c <= d",
			want:  []string{"a", ">=", "b", "AND", "c", "<=", "d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenizeSQL(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("tokenizeSQL() = %v, want %v", got, tt.want)
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("tokenizeSQL()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCheckBalancedQuotes(t *testing.T) {
	tests := []struct {
		input string
		quote rune
		want  bool
	}{
		{"'hello'", '\'', true},
		{"'hello", '\'', false},
		{"hello'", '\'', false},
		{"'hello' 'world'", '\'', true},
		{`"hello"`, '"', true},
		{`"hello`, '"', false},
		{"no quotes", '\'', true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := checkBalancedQuotes(tt.input, tt.quote)
			if got != tt.want {
				t.Errorf("checkBalancedQuotes(%q, %q) = %v, want %v", tt.input, string(tt.quote), got, tt.want)
			}
		})
	}
}

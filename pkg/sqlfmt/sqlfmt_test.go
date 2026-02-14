package sqlfmt

import (
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		opts  []Option
		want  string
	}{
		{
			name:  "simple select uppercase",
			input: "select * from users",
			opts:  []Option{WithUppercase()},
			want:  "SELECT *\nFROM users",
		},
		{
			name:  "select with where",
			input: "select id, name from users where active = true",
			opts:  []Option{WithUppercase()},
			want:  "SELECT id,\nname\nFROM users\nWHERE active = TRUE",
		},
		{
			name:  "select with join",
			input: "select u.id, o.amount from users u join orders o on u.id = o.user_id",
			opts:  []Option{WithUppercase()},
			want:  "SELECT u.id,\no.amount\nFROM users u\nJOIN orders o\nON u.id = o.user_id",
		},
		{
			name:  "insert",
			input: "insert into users (name, email) values ('John', 'john@example.com')",
			opts:  []Option{WithUppercase()},
			want:  "INSERT INTO users (name,\n  email)\nVALUES ('John',\n  'john@example.com')",
		},
		{
			name:  "update",
			input: "update users set name = 'Jane' where id = 1",
			opts:  []Option{WithUppercase()},
			want:  "UPDATE users\nSET name = 'Jane'\nWHERE id = 1",
		},
		{
			name:  "delete",
			input: "delete from users where id = 1",
			opts:  []Option{WithUppercase()},
			want:  "DELETE\nFROM users\nWHERE id = 1",
		},
		{
			name:  "preserve lowercase",
			input: "select * from users",
			opts:  nil,
			want:  "select *\nfrom users",
		},
		{
			name:  "order by",
			input: "select * from users order by name asc",
			opts:  []Option{WithUppercase()},
			want:  "SELECT *\nFROM users\nORDER BY name ASC",
		},
		{
			name:  "group by",
			input: "select count(*), status from orders group by status",
			opts:  []Option{WithUppercase()},
			want:  "SELECT COUNT(*),\nstatus\nFROM orders\nGROUP BY status",
		},
		{
			name:  "having",
			input: "select status, count(*) from orders group by status having count(*) > 10",
			opts:  []Option{WithUppercase()},
			want:  "SELECT status,\nCOUNT(*)\nFROM orders\nGROUP BY status\nHAVING COUNT(*) > 10",
		},
		{
			name:  "limit offset",
			input: "select * from users limit 10 offset 20",
			opts:  []Option{WithUppercase()},
			want:  "SELECT *\nFROM users\nLIMIT 10\nOFFSET 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Format(tt.input, tt.opts...)
			if got != tt.want {
				t.Errorf("Format() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestMinify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple",
			input: "SELECT * \nFROM users \nWHERE id = 1",
			want:  "SELECT * FROM users WHERE id = 1",
		},
		{
			name:  "extra whitespace",
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
			got := Minify(tt.input)
			if got != tt.want {
				t.Errorf("Minify() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValid bool
	}{
		{"valid select", "SELECT * FROM users", true},
		{"valid insert", "INSERT INTO users (name) VALUES ('John')", true},
		{"valid update", "UPDATE users SET name = 'Jane' WHERE id = 1", true},
		{"valid delete", "DELETE FROM users WHERE id = 1", true},
		{"valid create", "CREATE TABLE users (id INT PRIMARY KEY)", true},
		{"unbalanced open", "SELECT * FROM users WHERE (id = 1", false},
		{"unbalanced close", "SELECT * FROM users WHERE id = 1)", false},
		{"unbalanced quotes", "SELECT * FROM users WHERE name = 'John", false},
		{"invalid start", "INVALID FROM users", false},
		{"empty", "", false},
		{"comment", "-- this is a comment", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(tt.input)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate(%q).Valid = %v, want %v (error: %s)", tt.input, result.Valid, tt.wantValid, result.Error)
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
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			if got := IsKeyword(tt.token); got != tt.want {
				t.Errorf("IsKeyword(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}

func TestNeedsSpace(t *testing.T) {
	tests := []struct {
		prev, curr string
		want       bool
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
			if got := NeedsSpace(tt.prev, tt.curr); got != tt.want {
				t.Errorf("NeedsSpace(%q, %q) = %v, want %v", tt.prev, tt.curr, got, tt.want)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"simple", "SELECT * FROM users", []string{"SELECT", "*", "FROM", "users"}},
		{"operators", "id = 1", []string{"id", "=", "1"}},
		{"string literal", "name = 'John'", []string{"name", "=", "'John'"}},
		{"parens", "(id, name)", []string{"(", "id", ",", "name", ")"}},
		{"comparison", "a >= b AND c <= d", []string{"a", ">=", "b", "AND", "c", "<=", "d"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Tokenize(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("Tokenize() = %v, want %v", got, tt.want)
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Tokenize()[%d] = %q, want %q", i, got[i], tt.want[i])
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
			if got := CheckBalancedQuotes(tt.input, tt.quote); got != tt.want {
				t.Errorf("CheckBalancedQuotes(%q, %q) = %v, want %v", tt.input, string(tt.quote), got, tt.want)
			}
		})
	}
}

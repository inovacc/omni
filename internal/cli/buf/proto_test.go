package buf

import (
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantTokens int
		checkFunc  func([]Token) bool
	}{
		{
			name:       "simple syntax",
			input:      `syntax = "proto3";`,
			wantTokens: 5, // syntax, =, "proto3", ;, EOF
			checkFunc: func(tokens []Token) bool {
				return tokens[0].Value == "syntax" &&
					tokens[0].Type == TokenKeyword
			},
		},
		{
			name:       "package declaration",
			input:      `package test.v1;`,
			wantTokens: 5, // package, test, ., v1, ;, EOF
			checkFunc: func(tokens []Token) bool {
				return tokens[0].Value == "package" &&
					tokens[0].Type == TokenKeyword
			},
		},
		{
			name:       "message definition",
			input:      `message User { string name = 1; }`,
			wantTokens: 11,
			checkFunc: func(tokens []Token) bool {
				return tokens[0].Value == "message" &&
					tokens[0].Type == TokenKeyword
			},
		},
		{
			name:       "comment handling",
			input:      "// This is a comment\nsyntax = \"proto3\";",
			wantTokens: 7,
			checkFunc: func(tokens []Token) bool {
				hasComment := false

				for _, tok := range tokens {
					if tok.Type == TokenComment {
						hasComment = true
						break
					}
				}

				return hasComment
			},
		},
		{
			name:       "number token",
			input:      `int32 id = 123;`,
			wantTokens: 6,
			checkFunc: func(tokens []Token) bool {
				for _, tok := range tokens {
					if tok.Type == TokenNumber && tok.Value == "123" {
						return true
					}
				}

				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens := lexer.Tokenize()

			// Filter out whitespace and newlines for count
			filtered := filterTokens(tokens)

			if len(filtered) < tt.wantTokens-1 {
				t.Errorf("Lexer got %d tokens, want at least %d", len(filtered), tt.wantTokens-1)
			}

			if tt.checkFunc != nil && !tt.checkFunc(tokens) {
				t.Errorf("Lexer token check failed for input: %s", tt.input)
			}
		})
	}
}

func filterTokens(tokens []Token) []Token {
	var result []Token

	for _, t := range tokens {
		if t.Type != TokenWhitespace && t.Type != TokenNewline {
			result = append(result, t)
		}
	}

	return result
}

func TestParseProtoFile(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(*ProtoFile) bool
	}{
		{
			name: "minimal proto file",
			input: `syntax = "proto3";
package test;
`,
			wantErr: false,
			checkFunc: func(pf *ProtoFile) bool {
				return pf.Syntax == "proto3" && pf.Package == "test"
			},
		},
		{
			name: "proto with message",
			input: `syntax = "proto3";
package test;

message User {
  string name = 1;
  int32 age = 2;
}
`,
			wantErr: false,
			checkFunc: func(pf *ProtoFile) bool {
				return len(pf.Messages) == 1 &&
					pf.Messages[0].Name == "User" &&
					len(pf.Messages[0].Fields) == 2
			},
		},
		{
			name: "proto with enum",
			input: `syntax = "proto3";
package test;

enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}
`,
			wantErr: false,
			checkFunc: func(pf *ProtoFile) bool {
				return len(pf.Enums) == 1 &&
					pf.Enums[0].Name == "Status" &&
					len(pf.Enums[0].Values) == 3
			},
		},
		{
			name: "proto with service",
			input: `syntax = "proto3";
package test;

message Request {}
message Response {}

service UserService {
  rpc GetUser(Request) returns (Response);
}
`,
			wantErr: false,
			checkFunc: func(pf *ProtoFile) bool {
				return len(pf.Services) == 1 &&
					pf.Services[0].Name == "UserService" &&
					len(pf.Services[0].Methods) == 1
			},
		},
		{
			name: "proto with import",
			input: `syntax = "proto3";
package test;

import "google/protobuf/timestamp.proto";
import public "other.proto";
`,
			wantErr: false,
			checkFunc: func(pf *ProtoFile) bool {
				return len(pf.Imports) == 2 &&
					pf.Imports[0].Path == "google/protobuf/timestamp.proto"
			},
		},
		{
			name: "proto with options",
			input: `syntax = "proto3";
package test;

option go_package = "github.com/test/pkg";
option java_package = "com.test.pkg";
`,
			wantErr: false,
			checkFunc: func(pf *ProtoFile) bool {
				return len(pf.Options) == 2
			},
		},
		{
			name: "proto with nested message",
			input: `syntax = "proto3";
package test;

message Outer {
  message Inner {
    string value = 1;
  }
  Inner inner = 1;
}
`,
			wantErr: false,
			checkFunc: func(pf *ProtoFile) bool {
				return len(pf.Messages) == 1 &&
					len(pf.Messages[0].Nested) == 1 &&
					pf.Messages[0].Nested[0].Name == "Inner"
			},
		},
		{
			name: "proto with streaming rpc",
			input: `syntax = "proto3";
package test;

message Request {}
message Response {}

service StreamService {
  rpc ClientStream(stream Request) returns (Response);
  rpc ServerStream(Request) returns (stream Response);
  rpc BidiStream(stream Request) returns (stream Response);
}
`,
			wantErr: false,
			checkFunc: func(pf *ProtoFile) bool {
				if len(pf.Services) != 1 || len(pf.Services[0].Methods) != 3 {
					return false
				}

				methods := pf.Services[0].Methods

				return methods[0].ClientStreaming &&
					!methods[0].ServerStreaming &&
					!methods[1].ClientStreaming &&
					methods[1].ServerStreaming &&
					methods[2].ClientStreaming &&
					methods[2].ServerStreaming
			},
		},
		{
			name: "proto with map field",
			input: `syntax = "proto3";
package test;

message MapMessage {
  map<string, int32> values = 1;
}
`,
			wantErr: false,
			checkFunc: func(pf *ProtoFile) bool {
				return len(pf.Messages) == 1 &&
					len(pf.Messages[0].Fields) == 1
			},
		},
		{
			name: "proto with repeated field",
			input: `syntax = "proto3";
package test;

message ListMessage {
  repeated string items = 1;
}
`,
			wantErr: false,
			checkFunc: func(pf *ProtoFile) bool {
				return len(pf.Messages) == 1 &&
					len(pf.Messages[0].Fields) == 1 &&
					pf.Messages[0].Fields[0].Label == "repeated"
			},
		},
		{
			name: "proto with optional field",
			input: `syntax = "proto3";
package test;

message OptionalMessage {
  optional string name = 1;
}
`,
			wantErr: false,
			checkFunc: func(pf *ProtoFile) bool {
				return len(pf.Messages) == 1 &&
					len(pf.Messages[0].Fields) == 1 &&
					pf.Messages[0].Fields[0].Label == "optional"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf, err := ParseProtoFile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProtoFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil && !tt.checkFunc(pf) {
				t.Errorf("ParseProtoFile() check failed for: %s\nGot: %+v", tt.name, pf)
			}
		})
	}
}

func TestParseComplexProto(t *testing.T) {
	input := `syntax = "proto3";

package api.v1;

import "google/protobuf/timestamp.proto";

option go_package = "github.com/test/api/v1";

// User represents a user in the system
message User {
  string id = 1;
  string name = 2;
  string email = 3;
  google.protobuf.Timestamp created_at = 4;

  enum Role {
    ROLE_UNSPECIFIED = 0;
    ROLE_USER = 1;
    ROLE_ADMIN = 2;
  }

  Role role = 5;
  repeated string tags = 6;
  map<string, string> metadata = 7;
}

message CreateUserRequest {
  User user = 1;
}

message CreateUserResponse {
  User user = 1;
}

message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  User user = 1;
}

// UserService provides user management
service UserService {
  // CreateUser creates a new user
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);

  // GetUser retrieves a user by ID
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
`

	pf, err := ParseProtoFile(input)
	if err != nil {
		t.Fatalf("ParseProtoFile() error = %v", err)
	}

	if pf.Syntax != "proto3" {
		t.Errorf("Syntax = %s, want proto3", pf.Syntax)
	}

	if pf.Package != "api.v1" {
		t.Errorf("Package = %s, want api.v1", pf.Package)
	}

	if len(pf.Imports) != 1 {
		t.Errorf("Imports count = %d, want 1", len(pf.Imports))
	}

	if len(pf.Options) != 1 {
		t.Errorf("Options count = %d, want 1", len(pf.Options))
	}

	if len(pf.Messages) != 5 {
		t.Errorf("Messages count = %d, want 5", len(pf.Messages))
	}

	if len(pf.Services) != 1 {
		t.Errorf("Services count = %d, want 1", len(pf.Services))
	}

	if len(pf.Services[0].Methods) != 2 {
		t.Errorf("Service methods count = %d, want 2", len(pf.Services[0].Methods))
	}

	// Check User message
	var userMsg *ProtoMessage

	for i := range pf.Messages {
		if pf.Messages[i].Name == "User" {
			userMsg = &pf.Messages[i]
			break
		}
	}

	if userMsg == nil {
		t.Fatal("User message not found")
	}

	if len(userMsg.Fields) < 6 {
		t.Errorf("User fields count = %d, want at least 6", len(userMsg.Fields))
	}

	if len(userMsg.Enums) != 1 {
		t.Errorf("User nested enums count = %d, want 1", len(userMsg.Enums))
	}
}

func TestTokenTypes(t *testing.T) {
	input := `syntax = "proto3";
// comment
message Test {
  int32 id = 1;
}`

	lexer := NewLexer(input)
	tokens := lexer.Tokenize()

	typeFound := make(map[TokenType]bool)
	for _, tok := range tokens {
		typeFound[tok.Type] = true
	}

	expectedTypes := []TokenType{
		TokenKeyword,
		TokenSymbol,
		TokenString,
		TokenComment,
		TokenIdent,
		TokenNumber,
		TokenEOF,
	}

	for _, et := range expectedTypes {
		if !typeFound[et] {
			t.Errorf("Token type %d not found in tokens", et)
		}
	}
}

func TestParseProtoWithOneof(t *testing.T) {
	input := `syntax = "proto3";
package test;

message Example {
  oneof choice {
    string text = 1;
    int32 number = 2;
  }
}
`

	pf, err := ParseProtoFile(input)
	if err != nil {
		t.Fatalf("ParseProtoFile() error = %v", err)
	}

	if len(pf.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(pf.Messages))
	}
}

func TestParseProtoWithReserved(t *testing.T) {
	input := `syntax = "proto3";
package test;

message Example {
  reserved 2, 15, 9 to 11;
  reserved "foo", "bar";
  string name = 1;
}
`

	pf, err := ParseProtoFile(input)
	if err != nil {
		t.Fatalf("ParseProtoFile() error = %v", err)
	}

	if len(pf.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(pf.Messages))
	}

	if len(pf.Messages[0].Reserved) == 0 {
		t.Error("Expected reserved fields to be parsed")
	}
}

func TestParseProtoWithMessageOptions(t *testing.T) {
	input := `syntax = "proto3";
package test;

message Example {
  option deprecated = true;
  string name = 1;
}
`

	pf, err := ParseProtoFile(input)
	if err != nil {
		t.Fatalf("ParseProtoFile() error = %v", err)
	}

	if len(pf.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(pf.Messages))
	}
}

func TestParseProtoWithEnumOptions(t *testing.T) {
	input := `syntax = "proto3";
package test;

enum Status {
  option allow_alias = true;
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  ACTIVE = 1;
}
`

	pf, err := ParseProtoFile(input)
	if err != nil {
		t.Fatalf("ParseProtoFile() error = %v", err)
	}

	if len(pf.Enums) != 1 {
		t.Errorf("Expected 1 enum, got %d", len(pf.Enums))
	}
}

func TestParseProtoWithServiceOptions(t *testing.T) {
	input := `syntax = "proto3";
package test;

message Request {}
message Response {}

service TestService {
  option deprecated = true;
  rpc Get(Request) returns (Response) {
    option idempotency_level = NO_SIDE_EFFECTS;
  }
}
`

	pf, err := ParseProtoFile(input)
	if err != nil {
		t.Fatalf("ParseProtoFile() error = %v", err)
	}

	if len(pf.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(pf.Services))
	}
}

func TestParseProtoPublicImport(t *testing.T) {
	input := `syntax = "proto3";
package test;

import "other.proto";
import public "public.proto";
import weak "weak.proto";
`

	pf, err := ParseProtoFile(input)
	if err != nil {
		t.Fatalf("ParseProtoFile() error = %v", err)
	}

	if len(pf.Imports) != 3 {
		t.Errorf("Expected 3 imports, got %d", len(pf.Imports))
	}
}

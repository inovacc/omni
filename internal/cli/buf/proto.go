package buf

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// TokenType represents the type of a proto token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenString
	TokenNumber
	TokenKeyword
	TokenSymbol
	TokenComment
	TokenWhitespace
	TokenNewline
)

// Token represents a proto token
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// ProtoFile represents a parsed proto file
type ProtoFile struct {
	Syntax   string
	Package  string
	Options  []ProtoOption
	Imports  []ProtoImport
	Messages []ProtoMessage
	Enums    []ProtoEnum
	Services []ProtoService
	Comments []ProtoComment
}

// ProtoOption represents a proto option
type ProtoOption struct {
	Name  string
	Value string
	Line  int
}

// ProtoImport represents a proto import
type ProtoImport struct {
	Path   string
	Public bool
	Weak   bool
	Line   int
}

// ProtoMessage represents a proto message
type ProtoMessage struct {
	Name     string
	Fields   []ProtoField
	Nested   []ProtoMessage
	Enums    []ProtoEnum
	Options  []ProtoOption
	Reserved []string
	Line     int
	Comments []string
}

// ProtoField represents a proto field
type ProtoField struct {
	Name     string
	Type     string
	Number   int
	Label    string // optional, required, repeated
	Options  []ProtoOption
	Line     int
	Comments []string
}

// ProtoEnum represents a proto enum
type ProtoEnum struct {
	Name    string
	Values  []ProtoEnumValue
	Options []ProtoOption
	Line    int
}

// ProtoEnumValue represents a proto enum value
type ProtoEnumValue struct {
	Name    string
	Number  int
	Options []ProtoOption
	Line    int
}

// ProtoService represents a proto service
type ProtoService struct {
	Name    string
	Methods []ProtoMethod
	Options []ProtoOption
	Line    int
}

// ProtoMethod represents a proto RPC method
type ProtoMethod struct {
	Name            string
	InputType       string
	OutputType      string
	ClientStreaming bool
	ServerStreaming bool
	Options         []ProtoOption
	Line            int
}

// ProtoComment represents a comment in the proto file
type ProtoComment struct {
	Text string
	Line int
}

// Lexer tokenizes proto source
type Lexer struct {
	input  []rune
	pos    int
	line   int
	column int
	tokens []Token
}

// Keywords in proto3
var protoKeywords = map[string]bool{
	"syntax": true, "import": true, "weak": true, "public": true,
	"package": true, "option": true, "message": true, "enum": true,
	"service": true, "rpc": true, "returns": true, "stream": true,
	"optional": true, "repeated": true, "required": true, "reserved": true,
	"oneof": true, "map": true, "extensions": true, "extend": true,
	"true": true, "false": true, "to": true, "max": true,
	"double": true, "float": true, "int32": true, "int64": true,
	"uint32": true, "uint64": true, "sint32": true, "sint64": true,
	"fixed32": true, "fixed64": true, "sfixed32": true, "sfixed64": true,
	"bool": true, "string": true, "bytes": true,
}

// NewLexer creates a new lexer
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  []rune(input),
		line:   1,
		column: 1,
	}
}

// Tokenize returns all tokens from the input
func (l *Lexer) Tokenize() []Token {
	for l.pos < len(l.input) {
		l.nextToken()
	}

	l.tokens = append(l.tokens, Token{Type: TokenEOF, Line: l.line, Column: l.column})

	return l.tokens
}

func (l *Lexer) nextToken() {
	if l.pos >= len(l.input) {
		return
	}

	ch := l.input[l.pos]

	// Handle whitespace
	if ch == ' ' || ch == '\t' {
		l.scanWhitespace()
		return
	}

	// Handle newlines
	if ch == '\n' {
		l.tokens = append(l.tokens, Token{
			Type:   TokenNewline,
			Value:  "\n",
			Line:   l.line,
			Column: l.column,
		})
		l.pos++
		l.line++
		l.column = 1

		return
	}

	if ch == '\r' {
		l.pos++
		if l.pos < len(l.input) && l.input[l.pos] == '\n' {
			l.pos++
		}

		l.tokens = append(l.tokens, Token{
			Type:   TokenNewline,
			Value:  "\n",
			Line:   l.line,
			Column: l.column,
		})
		l.line++
		l.column = 1

		return
	}

	// Handle comments
	if ch == '/' && l.pos+1 < len(l.input) {
		next := l.input[l.pos+1]
		if next == '/' {
			l.scanLineComment()

			return
		}

		if next == '*' {
			l.scanBlockComment()

			return
		}
	}

	// Handle strings
	if ch == '"' || ch == '\'' {
		l.scanString(ch)

		return
	}

	// Handle numbers
	if unicode.IsDigit(ch) || (ch == '-' && l.pos+1 < len(l.input) && unicode.IsDigit(l.input[l.pos+1])) {
		l.scanNumber()

		return
	}

	// Handle identifiers and keywords
	if unicode.IsLetter(ch) || ch == '_' {
		l.scanIdentifier()

		return
	}

	// Handle symbols
	l.tokens = append(l.tokens, Token{
		Type:   TokenSymbol,
		Value:  string(ch),
		Line:   l.line,
		Column: l.column,
	})
	l.pos++
	l.column++
}

func (l *Lexer) scanWhitespace() {
	start := l.pos
	startCol := l.column

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch != ' ' && ch != '\t' {
			break
		}

		l.pos++
		l.column++
	}

	l.tokens = append(l.tokens, Token{
		Type:   TokenWhitespace,
		Value:  string(l.input[start:l.pos]),
		Line:   l.line,
		Column: startCol,
	})
}

func (l *Lexer) scanLineComment() {
	startCol := l.column
	l.pos += 2 // Skip //
	l.column += 2

	start := l.pos

	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
		l.column++
	}

	l.tokens = append(l.tokens, Token{
		Type:   TokenComment,
		Value:  "//" + string(l.input[start:l.pos]),
		Line:   l.line,
		Column: startCol,
	})
}

func (l *Lexer) scanBlockComment() {
	startCol := l.column
	startLine := l.line
	l.pos += 2 // Skip /*
	l.column += 2

	var builder strings.Builder
	builder.WriteString("/*")

	for l.pos < len(l.input) {
		if l.input[l.pos] == '*' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '/' {
			builder.WriteString("*/")

			l.pos += 2
			l.column += 2

			break
		}

		if l.input[l.pos] == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}

		builder.WriteRune(l.input[l.pos])
		l.pos++
	}

	l.tokens = append(l.tokens, Token{
		Type:   TokenComment,
		Value:  builder.String(),
		Line:   startLine,
		Column: startCol,
	})
}

func (l *Lexer) scanString(quote rune) {
	startCol := l.column
	l.pos++ // Skip opening quote
	l.column++

	var builder strings.Builder

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		if ch == quote {
			l.pos++
			l.column++

			break
		}

		if ch == '\\' && l.pos+1 < len(l.input) {
			builder.WriteRune(ch)

			l.pos++
			l.column++
			builder.WriteRune(l.input[l.pos])
			l.pos++
			l.column++

			continue
		}

		builder.WriteRune(ch)

		l.pos++
		l.column++
	}

	l.tokens = append(l.tokens, Token{
		Type:   TokenString,
		Value:  builder.String(),
		Line:   l.line,
		Column: startCol,
	})
}

func (l *Lexer) scanNumber() {
	startCol := l.column
	start := l.pos

	if l.input[l.pos] == '-' {
		l.pos++
		l.column++
	}

	// Handle hex
	if l.pos+1 < len(l.input) && l.input[l.pos] == '0' &&
		(l.input[l.pos+1] == 'x' || l.input[l.pos+1] == 'X') {
		l.pos += 2
		l.column += 2

		for l.pos < len(l.input) {
			ch := l.input[l.pos]
			if !unicode.IsDigit(ch) && (ch < 'a' || ch > 'f') && (ch < 'A' || ch > 'F') {
				break
			}

			l.pos++
			l.column++
		}
	} else {
		// Decimal
		for l.pos < len(l.input) && unicode.IsDigit(l.input[l.pos]) {
			l.pos++
			l.column++
		}

		// Handle decimal point
		if l.pos < len(l.input) && l.input[l.pos] == '.' {
			l.pos++
			l.column++

			for l.pos < len(l.input) && unicode.IsDigit(l.input[l.pos]) {
				l.pos++
				l.column++
			}
		}

		// Handle exponent
		if l.pos < len(l.input) && (l.input[l.pos] == 'e' || l.input[l.pos] == 'E') {
			l.pos++
			l.column++

			if l.pos < len(l.input) && (l.input[l.pos] == '+' || l.input[l.pos] == '-') {
				l.pos++
				l.column++
			}

			for l.pos < len(l.input) && unicode.IsDigit(l.input[l.pos]) {
				l.pos++
				l.column++
			}
		}
	}

	l.tokens = append(l.tokens, Token{
		Type:   TokenNumber,
		Value:  string(l.input[start:l.pos]),
		Line:   l.line,
		Column: startCol,
	})
}

func (l *Lexer) scanIdentifier() {
	startCol := l.column
	start := l.pos

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' && ch != '.' {
			break
		}

		l.pos++
		l.column++
	}

	value := string(l.input[start:l.pos])
	tokenType := TokenIdent

	if protoKeywords[value] {
		tokenType = TokenKeyword
	}

	l.tokens = append(l.tokens, Token{
		Type:   tokenType,
		Value:  value,
		Line:   l.line,
		Column: startCol,
	})
}

// Parser parses proto tokens into a ProtoFile
type Parser struct {
	tokens []Token
	pos    int
}

// NewParser creates a new parser
func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

// Parse parses the tokens into a ProtoFile
func (p *Parser) Parse() (*ProtoFile, error) {
	file := &ProtoFile{}

	for !p.isAtEnd() {
		if err := p.parseTopLevel(file); err != nil {
			return nil, err
		}
	}

	return file, nil
}

func (p *Parser) parseTopLevel(file *ProtoFile) error {
	p.skipWhitespaceAndComments()

	if p.isAtEnd() {
		return nil
	}

	token := p.current()

	switch token.Value {
	case "syntax":
		return p.parseSyntax(file)
	case "package":
		return p.parsePackage(file)
	case "import":
		return p.parseImport(file)
	case "option":
		return p.parseOption(file)
	case "message":
		return p.parseMessage(file)
	case "enum":
		p.parseEnum(file)

		return nil
	case "service":
		return p.parseService(file)
	default:
		// Skip unknown tokens
		p.advance()

		return nil
	}
}

func (p *Parser) parseSyntax(file *ProtoFile) error {
	p.advance() // skip "syntax"
	p.skipWhitespaceAndComments()

	if !p.match(TokenSymbol, "=") {
		return fmt.Errorf("expected '=' after syntax at line %d", p.current().Line)
	}

	p.skipWhitespaceAndComments()

	if p.current().Type != TokenString {
		return fmt.Errorf("expected string after '=' at line %d", p.current().Line)
	}

	file.Syntax = p.current().Value
	p.advance()
	p.skipWhitespaceAndComments()
	p.match(TokenSymbol, ";")

	return nil
}

func (p *Parser) parsePackage(file *ProtoFile) error {
	p.advance() // skip "package"
	p.skipWhitespaceAndComments()

	if p.current().Type != TokenIdent {
		return fmt.Errorf("expected package name at line %d", p.current().Line)
	}

	file.Package = p.current().Value
	p.advance()
	p.skipWhitespaceAndComments()
	p.match(TokenSymbol, ";")

	return nil
}

func (p *Parser) parseImport(file *ProtoFile) error {
	line := p.current().Line
	p.advance() // skip "import"
	p.skipWhitespaceAndComments()

	imp := ProtoImport{Line: line}

	if p.current().Value == "public" {
		imp.Public = true

		p.advance()
		p.skipWhitespaceAndComments()
	} else if p.current().Value == "weak" {
		imp.Weak = true

		p.advance()
		p.skipWhitespaceAndComments()
	}

	if p.current().Type != TokenString {
		return fmt.Errorf("expected import path at line %d", p.current().Line)
	}

	imp.Path = p.current().Value
	p.advance()
	p.skipWhitespaceAndComments()
	p.match(TokenSymbol, ";")

	file.Imports = append(file.Imports, imp)

	return nil
}

func (p *Parser) parseOption(file *ProtoFile) error {
	line := p.current().Line
	p.advance() // skip "option"
	p.skipWhitespaceAndComments()

	opt := ProtoOption{Line: line}

	// Parse option name
	var name strings.Builder

	for !p.isAtEnd() && !p.check(TokenSymbol, "=") {
		if p.current().Type == TokenWhitespace || p.current().Type == TokenNewline {
			p.advance()

			continue
		}

		name.WriteString(p.current().Value)
		p.advance()
	}

	opt.Name = strings.TrimSpace(name.String())

	p.match(TokenSymbol, "=")
	p.skipWhitespaceAndComments()

	// Parse option value
	var value strings.Builder

	for !p.isAtEnd() && !p.check(TokenSymbol, ";") {
		if p.current().Type == TokenWhitespace || p.current().Type == TokenNewline {
			p.advance()

			continue
		}

		value.WriteString(p.current().Value)
		p.advance()
	}

	opt.Value = strings.TrimSpace(value.String())

	p.match(TokenSymbol, ";")

	file.Options = append(file.Options, opt)

	return nil
}

func (p *Parser) parseMessage(file *ProtoFile) error {
	line := p.current().Line
	p.advance() // skip "message"
	p.skipWhitespaceAndComments()

	msg := ProtoMessage{Line: line}

	if p.current().Type != TokenIdent {
		return fmt.Errorf("expected message name at line %d", p.current().Line)
	}

	msg.Name = p.current().Value
	p.advance()
	p.skipWhitespaceAndComments()

	if !p.match(TokenSymbol, "{") {
		return fmt.Errorf("expected '{' after message name at line %d", p.current().Line)
	}

	// Parse message body
	for !p.isAtEnd() && !p.check(TokenSymbol, "}") {
		p.skipWhitespaceAndComments()

		if p.check(TokenSymbol, "}") {
			break
		}

		if err := p.parseMessageBody(&msg); err != nil {
			return err
		}
	}

	p.match(TokenSymbol, "}")

	file.Messages = append(file.Messages, msg)

	return nil
}

func (p *Parser) parseMessageBody(msg *ProtoMessage) error {
	token := p.current()

	switch token.Value {
	case "message":
		// Nested message
		nested := ProtoMessage{Line: token.Line}

		p.advance()
		p.skipWhitespaceAndComments()

		nested.Name = p.current().Value
		p.advance()
		p.skipWhitespaceAndComments()
		p.match(TokenSymbol, "{")

		for !p.isAtEnd() && !p.check(TokenSymbol, "}") {
			p.skipWhitespaceAndComments()

			if p.check(TokenSymbol, "}") {
				break
			}

			if err := p.parseMessageBody(&nested); err != nil {
				return err
			}
		}

		p.match(TokenSymbol, "}")

		msg.Nested = append(msg.Nested, nested)

	case "enum":
		enum := p.parseEnumInline()
		msg.Enums = append(msg.Enums, *enum)

	case "option":
		opt := p.parseOptionInline()
		msg.Options = append(msg.Options, *opt)

	case "reserved":
		p.parseReserved(msg)

	case "oneof":
		p.parseOneof(msg)

	default:
		// Parse field
		if err := p.parseField(msg); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) parseField(msg *ProtoMessage) error {
	line := p.current().Line
	field := ProtoField{Line: line}

	// Check for label (optional, required, repeated)
	if p.current().Value == "optional" || p.current().Value == "required" || p.current().Value == "repeated" {
		field.Label = p.current().Value
		p.advance()
		p.skipWhitespaceAndComments()
	}

	// Parse type
	field.Type = p.current().Value
	p.advance()
	p.skipWhitespaceAndComments()

	// Handle map type
	if field.Type == "map" {
		p.match(TokenSymbol, "<")
		p.skipWhitespaceAndComments()

		var mapType strings.Builder
		mapType.WriteString("map<")

		for !p.isAtEnd() && !p.check(TokenSymbol, ">") {
			mapType.WriteString(p.current().Value)
			p.advance()
		}

		mapType.WriteString(">")
		p.match(TokenSymbol, ">")
		p.skipWhitespaceAndComments()

		field.Type = mapType.String()
	}

	// Parse name
	if p.current().Type != TokenIdent {
		p.advance()

		return nil
	}

	field.Name = p.current().Value
	p.advance()
	p.skipWhitespaceAndComments()

	// Parse =
	if !p.match(TokenSymbol, "=") {
		return nil
	}

	p.skipWhitespaceAndComments()

	// Parse number
	if p.current().Type == TokenNumber {
		field.Number, _ = strconv.Atoi(p.current().Value)
		p.advance()
	}

	// Skip to end of field
	for !p.isAtEnd() && !p.check(TokenSymbol, ";") && !p.check(TokenSymbol, "}") {
		p.advance()
	}

	p.match(TokenSymbol, ";")

	msg.Fields = append(msg.Fields, field)

	return nil
}

func (p *Parser) parseEnumInline() *ProtoEnum {
	line := p.current().Line
	p.advance() // skip "enum"
	p.skipWhitespaceAndComments()

	enum := &ProtoEnum{Line: line}
	enum.Name = p.current().Value
	p.advance()
	p.skipWhitespaceAndComments()
	p.match(TokenSymbol, "{")

	for !p.isAtEnd() && !p.check(TokenSymbol, "}") {
		p.skipWhitespaceAndComments()

		if p.check(TokenSymbol, "}") {
			break
		}

		if p.current().Value == "option" {
			opt := p.parseOptionInline()
			enum.Options = append(enum.Options, *opt)

			continue
		}

		// Parse enum value
		value := ProtoEnumValue{Line: p.current().Line}
		value.Name = p.current().Value
		p.advance()
		p.skipWhitespaceAndComments()
		p.match(TokenSymbol, "=")
		p.skipWhitespaceAndComments()

		if p.current().Type == TokenNumber {
			value.Number, _ = strconv.Atoi(p.current().Value)
			p.advance()
		}

		// Skip options and semicolon
		for !p.isAtEnd() && !p.check(TokenSymbol, ";") {
			p.advance()
		}

		p.match(TokenSymbol, ";")

		enum.Values = append(enum.Values, value)
	}

	p.match(TokenSymbol, "}")

	return enum
}

func (p *Parser) parseOptionInline() *ProtoOption {
	line := p.current().Line
	p.advance() // skip "option"
	p.skipWhitespaceAndComments()

	opt := &ProtoOption{Line: line}

	var name strings.Builder

	for !p.isAtEnd() && !p.check(TokenSymbol, "=") {
		if p.current().Type == TokenWhitespace {
			p.advance()

			continue
		}

		name.WriteString(p.current().Value)
		p.advance()
	}

	opt.Name = strings.TrimSpace(name.String())

	p.match(TokenSymbol, "=")
	p.skipWhitespaceAndComments()

	var value strings.Builder

	for !p.isAtEnd() && !p.check(TokenSymbol, ";") {
		value.WriteString(p.current().Value)
		p.advance()
	}

	opt.Value = strings.TrimSpace(value.String())

	p.match(TokenSymbol, ";")

	return opt
}

func (p *Parser) parseReserved(msg *ProtoMessage) {
	p.advance() // skip "reserved"

	for !p.isAtEnd() && !p.check(TokenSymbol, ";") {
		if p.current().Type == TokenString {
			msg.Reserved = append(msg.Reserved, p.current().Value)
		} else if p.current().Type == TokenNumber {
			msg.Reserved = append(msg.Reserved, p.current().Value)
		}

		p.advance()
	}

	p.match(TokenSymbol, ";")
}

func (p *Parser) parseOneof(msg *ProtoMessage) {
	p.advance() // skip "oneof"
	p.skipWhitespaceAndComments()
	p.advance() // skip name
	p.skipWhitespaceAndComments()
	p.match(TokenSymbol, "{")

	// Parse oneof body
	for !p.isAtEnd() && !p.check(TokenSymbol, "}") {
		p.skipWhitespaceAndComments()

		if p.check(TokenSymbol, "}") {
			break
		}

		_ = p.parseField(msg)
	}

	p.match(TokenSymbol, "}")
}

func (p *Parser) parseEnum(file *ProtoFile) {
	enum := p.parseEnumInline()
	file.Enums = append(file.Enums, *enum)
}

func (p *Parser) parseService(file *ProtoFile) error {
	line := p.current().Line
	p.advance() // skip "service"
	p.skipWhitespaceAndComments()

	svc := ProtoService{Line: line}
	svc.Name = p.current().Value
	p.advance()
	p.skipWhitespaceAndComments()
	p.match(TokenSymbol, "{")

	for !p.isAtEnd() && !p.check(TokenSymbol, "}") {
		p.skipWhitespaceAndComments()

		if p.check(TokenSymbol, "}") {
			break
		}

		if p.current().Value == "rpc" {
			method := p.parseRPC()
			svc.Methods = append(svc.Methods, *method)
		} else if p.current().Value == "option" {
			opt := p.parseOptionInline()
			svc.Options = append(svc.Options, *opt)
		} else {
			p.advance()
		}
	}

	p.match(TokenSymbol, "}")

	file.Services = append(file.Services, svc)

	return nil
}

func (p *Parser) parseRPC() *ProtoMethod {
	line := p.current().Line
	p.advance() // skip "rpc"
	p.skipWhitespaceAndComments()

	method := &ProtoMethod{Line: line}
	method.Name = p.current().Value
	p.advance()
	p.skipWhitespaceAndComments()
	p.match(TokenSymbol, "(")
	p.skipWhitespaceAndComments()

	// Check for client streaming
	if p.current().Value == "stream" {
		method.ClientStreaming = true

		p.advance()
		p.skipWhitespaceAndComments()
	}

	method.InputType = p.current().Value
	p.advance()
	p.skipWhitespaceAndComments()
	p.match(TokenSymbol, ")")
	p.skipWhitespaceAndComments()

	// Skip "returns"
	if p.current().Value == "returns" {
		p.advance()
	}

	p.skipWhitespaceAndComments()
	p.match(TokenSymbol, "(")
	p.skipWhitespaceAndComments()

	// Check for server streaming
	if p.current().Value == "stream" {
		method.ServerStreaming = true

		p.advance()
		p.skipWhitespaceAndComments()
	}

	method.OutputType = p.current().Value
	p.advance()
	p.skipWhitespaceAndComments()
	p.match(TokenSymbol, ")")
	p.skipWhitespaceAndComments()

	// Parse method body or semicolon
	if p.check(TokenSymbol, "{") {
		p.match(TokenSymbol, "{")

		for !p.isAtEnd() && !p.check(TokenSymbol, "}") {
			if p.current().Value == "option" {
				opt := p.parseOptionInline()
				method.Options = append(method.Options, *opt)
			} else {
				p.advance()
			}
		}

		p.match(TokenSymbol, "}")
	} else {
		p.match(TokenSymbol, ";")
	}

	return method
}

func (p *Parser) skipWhitespaceAndComments() {
	for !p.isAtEnd() {
		t := p.current()
		if t.Type == TokenWhitespace || t.Type == TokenNewline || t.Type == TokenComment {
			p.advance()
		} else {
			break
		}
	}
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}

	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	if !p.isAtEnd() {
		p.pos++
	}
}

func (p *Parser) isAtEnd() bool {
	return p.pos >= len(p.tokens) || p.tokens[p.pos].Type == TokenEOF
}

func (p *Parser) check(_ TokenType, value string) bool {
	if p.isAtEnd() {
		return false
	}

	return p.current().Type == TokenSymbol && p.current().Value == value
}

func (p *Parser) match(_ TokenType, value string) bool {
	if p.check(TokenSymbol, value) {
		p.advance()

		return true
	}

	return false
}

// ParseProtoFile parses a proto file from source
func ParseProtoFile(source string) (*ProtoFile, error) {
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()
	parser := NewParser(tokens)

	return parser.Parse()
}

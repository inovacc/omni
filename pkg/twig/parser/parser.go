package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/inovacc/omni/pkg/twig/models"
)

// Sentinel errors for parser operations
var (
	ErrEmptyInput       = errors.New("empty tree structure")
	ErrInvalidFormat    = errors.New("invalid tree format")
	ErrEmptyNodeName    = errors.New("empty node name")
	ErrInvalidStructure = errors.New("invalid tree structure")
)

// TreeParser defines the interface for parsing tree structures
type TreeParser interface {
	Parse(reader io.Reader) (*models.Node, error)
	ParseString(content string) (*models.Node, error)
}

// Parser parses tree format text into node structure
type Parser struct{}

// Ensure Parser implements TreeParser
var _ TreeParser = (*Parser)(nil)

// NewParser creates a new parser and returns a TreeParser interface
func NewParser() TreeParser {
	return &Parser{}
}

// Parse parses a tree format from a reader
func (p *Parser) Parse(reader io.Reader) (*models.Node, error) {
	scanner := bufio.NewScanner(reader)

	var root *models.Node

	var currentPath []*models.Node // Stack to track hierarchy

	var prevLevel = -1

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// The first non-empty line is the root
		if root == nil {
			name, comment := p.parseName(line)

			isDir := strings.HasSuffix(name, "/")
			if isDir {
				name = strings.TrimSuffix(name, "/")
			}

			root = models.NewNode(name, name, isDir)
			root.Comment = comment
			currentPath = append(currentPath, root)

			continue
		}

		// Parse the line
		level, name, comment, err := p.parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		isDir := strings.HasSuffix(name, "/")
		if isDir {
			name = strings.TrimSuffix(name, "/")
		}

		// Create node
		node := models.NewNode(name, "", isDir)
		node.Comment = comment

		// Adjust currentPath based on level
		if level > prevLevel {
			// Going deeper - do nothing, parent is at end of currentPath
		} else if level <= prevLevel {
			// Going back up or same level
			// Trim currentPath to the parent at this level
			currentPath = currentPath[:level+1]
		}

		// Get parent (last item in currentPath)
		if len(currentPath) == 0 {
			return nil, fmt.Errorf("line %d: %w", lineNum, ErrInvalidStructure)
		}

		parent := currentPath[len(currentPath)-1]
		parent.AddChild(node)

		// If this is a directory, add to currentPath for potential children
		if isDir {
			currentPath = append(currentPath, node)
		}

		prevLevel = level
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if root == nil {
		return nil, ErrEmptyInput
	}

	// Build full paths
	p.buildPaths(root, "")

	return root, nil
}

// Box-drawing glyphs used by the formatter when rendering a tree. These are
// multi-byte UTF-8 runes, so they MUST be compared with strings.HasPrefix
// (rune-correct) rather than fixed-width byte windows such as line[i:i+4].
const (
	indentPipe  = "│   " // ancestor continuation: pipe + 3 spaces (4 display columns)
	indentBlank = "    " // last-child continuation: 4 spaces
	connectorT  = "├── " // mid child connector
	connectorL  = "└── " // last child connector
	connectorTB = "├──"  // mid child connector without trailing space
	connectorLB = "└──"  // last child connector without trailing space
)

// parseLine parses a tree format line.
// Returns: level, name, comment, error.
//
// The level is the number of indentation units ("│   " or "    ") that precede
// the connector ("├── "/"└── "); the connector itself does not add a level. This
// mirrors the grammar emitted by the formatter so that formatter output round-trips
// back through the parser with correct nesting.
func (p *Parser) parseLine(line string) (int, string, string, error) {
	level := 0
	rest := line

	// Consume leading indentation units. Each unit is one nesting level.
	for {
		switch {
		case strings.HasPrefix(rest, indentPipe):
			level++
			rest = rest[len(indentPipe):]
		case strings.HasPrefix(rest, indentBlank):
			level++
			rest = rest[len(indentBlank):]
		default:
			goto connector
		}
	}

connector:
	// Consume the connector glyph that introduces this node's name.
	switch {
	case strings.HasPrefix(rest, connectorT):
		rest = rest[len(connectorT):]
	case strings.HasPrefix(rest, connectorL):
		rest = rest[len(connectorL):]
	case strings.HasPrefix(rest, connectorTB):
		rest = rest[len(connectorTB):]
	case strings.HasPrefix(rest, connectorLB):
		rest = rest[len(connectorLB):]
	}

	// Extract name and comment
	remaining := strings.TrimLeft(rest, " ")
	if remaining == "" {
		return 0, "", "", ErrEmptyNodeName
	}

	name, comment := p.parseName(remaining)

	return level, name, comment, nil
}

// parseName extracts name and optional comment from a line
func (p *Parser) parseName(line string) (string, string) {
	// Look for the comment marker
	parts := strings.SplitN(line, "#", 2)
	name := strings.TrimSpace(parts[0])
	comment := ""

	if len(parts) > 1 {
		comment = strings.TrimSpace(parts[1])
	}

	return name, comment
}

// buildPaths builds full paths for all nodes
func (p *Parser) buildPaths(node *models.Node, parentPath string) {
	if node == nil {
		return
	}

	// Build this node's path
	if parentPath == "" {
		node.Path = node.Name
	} else {
		node.Path = parentPath + "/" + node.Name
	}

	// Recursively build children paths
	for _, child := range node.Children {
		p.buildPaths(child, node.Path)
	}
}

// ParseString is a convenience method to parse from string
func (p *Parser) ParseString(content string) (*models.Node, error) {
	return p.Parse(strings.NewReader(content))
}

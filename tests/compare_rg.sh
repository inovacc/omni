#!/bin/bash
# Compare ripgrep (rg) vs omni rg behavior
# This script creates test fixtures and compares output between the two implementations

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
PASS=0
FAIL=0
SKIP=0

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Build omni if needed
OMNI="$PROJECT_ROOT/omni"
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    OMNI="$PROJECT_ROOT/omni.exe"
fi

if [[ ! -f "$OMNI" ]]; then
    echo "Building omni..."
    cd "$PROJECT_ROOT" && go build -o "$OMNI" .
fi

# Create temp directory for test fixtures
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

echo -e "${BLUE}=== ripgrep vs omni rg Comparison Test ===${NC}"
echo "Test directory: $TEST_DIR"
echo ""

# Create test fixtures
create_fixtures() {
    # Go files
    mkdir -p "$TEST_DIR/src/pkg"

    cat > "$TEST_DIR/src/main.go" << 'EOF'
package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("Hello, World!")
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    return nil
}
EOF

    cat > "$TEST_DIR/src/pkg/helper.go" << 'EOF'
package pkg

// Helper is a helper function
func Helper() string {
    return "helper"
}

// TODO: Add more helpers
// FIXME: This needs improvement
EOF

    cat > "$TEST_DIR/src/pkg/helper_test.go" << 'EOF'
package pkg

import "testing"

func TestHelper(t *testing.T) {
    result := Helper()
    if result != "helper" {
        t.Errorf("expected helper, got %s", result)
    }
}
EOF

    # JavaScript files
    mkdir -p "$TEST_DIR/src/js"
    cat > "$TEST_DIR/src/js/app.js" << 'EOF'
const express = require('express');

function main() {
    console.log("Hello from JavaScript");
}

// TODO: Add error handling
main();
EOF

    # Python files
    mkdir -p "$TEST_DIR/src/py"
    cat > "$TEST_DIR/src/py/app.py" << 'EOF'
#!/usr/bin/env python3

def main():
    print("Hello from Python")

# TODO: Add logging
if __name__ == "__main__":
    main()
EOF

    # Markdown
    cat > "$TEST_DIR/README.md" << 'EOF'
# Test Project

This is a test project for comparing rg implementations.

## Features

- Feature 1: Hello World
- Feature 2: Error handling

## TODO

- Add more tests
- Improve documentation
EOF

    # JSON config
    cat > "$TEST_DIR/config.json" << 'EOF'
{
    "name": "test-project",
    "version": "1.0.0",
    "hello": "world",
    "features": ["hello", "world"]
}
EOF

    # Hidden file
    cat > "$TEST_DIR/.hidden" << 'EOF'
This is a hidden file
secret: password123
EOF

    # Create .gitignore
    cat > "$TEST_DIR/.gitignore" << 'EOF'
*.log
node_modules/
dist/
.env
EOF

    # Create ignored files
    echo "debug info" > "$TEST_DIR/debug.log"
    mkdir -p "$TEST_DIR/node_modules/pkg"
    echo "module content" > "$TEST_DIR/node_modules/pkg/index.js"

    # Binary file (with null bytes)
    printf 'binary\x00content\x00here' > "$TEST_DIR/binary.bin"

    # File with special patterns
    cat > "$TEST_DIR/patterns.txt" << 'EOF'
hello world
Hello World
HELLO WORLD
hello-world
hello_world
func()
func(arg)
func(arg1, arg2)
test@example.com
192.168.1.1
EOF
}

# Compare function
compare_output() {
    local description="$1"
    local rg_cmd="$2"
    local omni_cmd="$3"
    local allow_diff="${4:-false}"

    # Run both commands
    rg_output=$(eval "$rg_cmd" 2>&1 || true)
    omni_output=$(eval "$omni_cmd" 2>&1 || true)

    # Normalize output (sort lines, trim whitespace)
    rg_normalized=$(echo "$rg_output" | sort | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    omni_normalized=$(echo "$omni_output" | sort | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

    if [[ "$rg_normalized" == "$omni_normalized" ]]; then
        echo -e "${GREEN}[PASS]${NC} $description"
        ((PASS++))
    elif [[ "$allow_diff" == "true" ]]; then
        echo -e "${YELLOW}[DIFF]${NC} $description (allowed difference)"
        echo "  rg output lines: $(echo "$rg_output" | wc -l)"
        echo "  omni output lines: $(echo "$omni_output" | wc -l)"
        ((SKIP++))
    else
        echo -e "${RED}[FAIL]${NC} $description"
        echo "  Command (rg): $rg_cmd"
        echo "  Command (omni): $omni_cmd"
        echo "  --- rg output ---"
        echo "$rg_output" | head -10
        echo "  --- omni output ---"
        echo "$omni_output" | head -10
        ((FAIL++))
    fi
}

# Compare match count only
compare_count() {
    local description="$1"
    local rg_cmd="$2"
    local omni_cmd="$3"

    rg_count=$(eval "$rg_cmd" 2>&1 | wc -l || echo "0")
    omni_count=$(eval "$omni_cmd" 2>&1 | wc -l || echo "0")

    if [[ "$rg_count" == "$omni_count" ]]; then
        echo -e "${GREEN}[PASS]${NC} $description (both: $rg_count lines)"
        ((PASS++))
    else
        echo -e "${RED}[FAIL]${NC} $description"
        echo "  rg: $rg_count lines, omni: $omni_count lines"
        ((FAIL++))
    fi
}

# Feature presence test
test_feature() {
    local description="$1"
    local omni_cmd="$2"
    local expected="$3"

    output=$(eval "$omni_cmd" 2>&1 || true)

    if echo "$output" | grep -q "$expected"; then
        echo -e "${GREEN}[PASS]${NC} $description"
        ((PASS++))
    else
        echo -e "${RED}[FAIL]${NC} $description"
        echo "  Expected to find: $expected"
        echo "  Got: $(echo "$output" | head -5)"
        ((FAIL++))
    fi
}

# Run tests
run_tests() {
    cd "$TEST_DIR"

    echo -e "\n${BLUE}--- Basic Pattern Matching ---${NC}"

    compare_count "Simple pattern search" \
        "rg 'Hello' ." \
        "$OMNI rg 'Hello' ."

    compare_count "Case insensitive (-i)" \
        "rg -i 'hello' ." \
        "$OMNI rg -i 'hello' ."

    compare_count "Word boundary (-w)" \
        "rg -w 'main' ." \
        "$OMNI rg -w 'main' ."

    compare_count "Fixed string (-F)" \
        "rg -F 'func()' ." \
        "$OMNI rg -F 'func()' ."

    compare_count "Invert match (-v) single file" \
        "rg -v 'func' src/main.go" \
        "$OMNI rg -v 'func' src/main.go"

    echo -e "\n${BLUE}--- File Type Filtering ---${NC}"

    compare_count "Go files only (-t go)" \
        "rg -t go 'func' ." \
        "$OMNI rg -t go 'func' ."

    compare_count "JavaScript files (-t js)" \
        "rg -t js 'function' ." \
        "$OMNI rg -t js 'function' ."

    compare_count "Python files (-t py)" \
        "rg -t py 'def' ." \
        "$OMNI rg -t py 'def' ."

    compare_count "Markdown files (-t md)" \
        "rg -t md 'TODO' ." \
        "$OMNI rg -t md 'TODO' ."

    echo -e "\n${BLUE}--- Output Modes ---${NC}"

    compare_count "Count mode (-c)" \
        "rg -c 'TODO' ." \
        "$OMNI rg -c 'TODO' ."

    compare_count "Files with matches (-l)" \
        "rg -l 'Hello' ." \
        "$OMNI rg -l 'Hello' ."

    compare_count "Line numbers (-n)" \
        "rg -n 'main' src/main.go" \
        "$OMNI rg -n 'main' src/main.go"

    echo -e "\n${BLUE}--- Context Lines ---${NC}"

    compare_count "After context (-A 2)" \
        "rg -A 2 'import' src/main.go" \
        "$OMNI rg -A 2 'import' src/main.go"

    compare_count "Before context (-B 2)" \
        "rg -B 2 'run()' src/main.go" \
        "$OMNI rg -B 2 'run()' src/main.go"

    compare_count "Context (-C 1)" \
        "rg -C 1 'Println' src/main.go" \
        "$OMNI rg -C 1 'Println' src/main.go"

    echo -e "\n${BLUE}--- Glob Patterns ---${NC}"

    compare_count "Include glob (-g '*.go')" \
        "rg -g '*.go' 'func' ." \
        "$OMNI rg -g '*.go' 'func' ."

    compare_count "Exclude glob (-g '!*_test.go')" \
        "rg -g '!*_test.go' -t go 'func' ." \
        "$OMNI rg -g '!*_test.go' -t go 'func' ."

    echo -e "\n${BLUE}--- Hidden Files & Gitignore ---${NC}"

    compare_count "Default (respects gitignore)" \
        "rg 'debug' ." \
        "$OMNI rg 'debug' ."

    # Note: Hidden file behavior may differ
    test_feature "Hidden files (--hidden)" \
        "$OMNI rg --hidden 'secret' ." \
        "secret"

    compare_count "No ignore (--no-ignore)" \
        "rg --no-ignore 'debug' ." \
        "$OMNI rg --no-ignore 'debug' ."

    echo -e "\n${BLUE}--- Binary Files ---${NC}"

    # Both should skip binary files by default
    compare_count "Skip binary files" \
        "rg 'binary' ." \
        "$OMNI rg 'binary' ."

    echo -e "\n${BLUE}--- Special Characters ---${NC}"

    compare_count "Email pattern" \
        "rg '@' patterns.txt" \
        "$OMNI rg '@' patterns.txt"

    compare_count "IP address pattern" \
        "rg -F '192.168' patterns.txt" \
        "$OMNI rg -F '192.168' patterns.txt"

    echo -e "\n${BLUE}--- Omni-specific Features ---${NC}"

    test_feature "JSON output (--json)" \
        "$OMNI rg --json 'Hello' ." \
        '"total_matches"'

    test_feature "Max depth (--max-depth)" \
        "$OMNI rg --max-depth 1 'func' ." \
        "func"

    echo -e "\n${BLUE}--- Edge Cases ---${NC}"

    compare_count "No matches" \
        "rg 'NONEXISTENT_PATTERN_12345' ." \
        "$OMNI rg 'NONEXISTENT_PATTERN_12345' ."

    compare_count "Empty pattern handling" \
        "rg '' . 2>&1 | head -1" \
        "$OMNI rg '' . 2>&1 | head -1"

    compare_count "Multiple paths" \
        "rg 'func' src/main.go src/pkg/helper.go" \
        "$OMNI rg 'func' src/main.go src/pkg/helper.go"
}

# Create fixtures and run tests
create_fixtures
run_tests

# Summary
echo ""
echo -e "${BLUE}=== Summary ===${NC}"
echo -e "${GREEN}Passed: $PASS${NC}"
echo -e "${RED}Failed: $FAIL${NC}"
echo -e "${YELLOW}Skipped/Diff: $SKIP${NC}"
echo ""

if [[ $FAIL -gt 0 ]]; then
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi

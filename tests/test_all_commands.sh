#!/bin/bash
#
# Comprehensive test script for omni - all commands and features
# Usage: ./tests/test_all_commands.sh [--verbose] [--category CATEGORY]
#
# Categories: core, file, text, system, archive, hash, encoding, data, util, id, all
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
SKIPPED=0
TOTAL=0

# Options
VERBOSE=false
CATEGORY="all"
OMNI="${OMNI:-./omni}"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --category|-c)
            CATEGORY="$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [--verbose] [--category CATEGORY]"
            echo "Categories: core, file, text, system, archive, hash, encoding, data, util, id, all"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Setup test directory
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
    ((TOTAL++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    if [ "$VERBOSE" = true ] && [ -n "$2" ]; then
        echo -e "${RED}       Error: $2${NC}"
    fi
    ((FAILED++))
    ((TOTAL++))
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    ((SKIPPED++))
    ((TOTAL++))
}

log_section() {
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Test helper - run command and check exit code
run_test() {
    local name="$1"
    local cmd="$2"
    local expected_exit="${3:-0}"

    if $VERBOSE; then
        echo -e "${BLUE}  Running:${NC} $cmd"
    fi

    set +e
    output=$(eval "$cmd" 2>&1)
    exit_code=$?
    set -e

    if [ "$exit_code" -eq "$expected_exit" ]; then
        log_pass "$name"
        return 0
    else
        log_fail "$name" "Exit code $exit_code (expected $expected_exit): $output"
        return 1
    fi
}

# Test helper - run command and check output contains string
run_test_contains() {
    local name="$1"
    local cmd="$2"
    local expected="$3"

    if $VERBOSE; then
        echo -e "${BLUE}  Running:${NC} $cmd"
    fi

    set +e
    output=$(eval "$cmd" 2>&1)
    exit_code=$?
    set -e

    if [ "$exit_code" -eq 0 ] && echo "$output" | grep -q "$expected"; then
        log_pass "$name"
        return 0
    else
        log_fail "$name" "Output does not contain '$expected': $output"
        return 1
    fi
}

# Test helper - run command and check exact output
run_test_exact() {
    local name="$1"
    local cmd="$2"
    local expected="$3"

    if $VERBOSE; then
        echo -e "${BLUE}  Running:${NC} $cmd"
    fi

    set +e
    output=$(eval "$cmd" 2>&1)
    exit_code=$?
    set -e

    if [ "$exit_code" -eq 0 ] && [ "$output" = "$expected" ]; then
        log_pass "$name"
        return 0
    else
        log_fail "$name" "Expected '$expected', got '$output'"
        return 1
    fi
}

# Create test fixtures
create_fixtures() {
    log_info "Creating test fixtures in $TEST_DIR"

    # Text files
    echo -e "Hello World\nFoo Bar\nBaz Qux" > "$TEST_DIR/test.txt"
    echo -e "line1\nline2\nline3\nline4\nline5" > "$TEST_DIR/lines.txt"
    echo -e "apple\nbanana\napple\ncherry\nbanana\nbanana" > "$TEST_DIR/duplicates.txt"
    echo -e "3\n1\n4\n1\n5\n9\n2\n6" > "$TEST_DIR/numbers.txt"
    echo -e "name,age,city\nAlice,30,NYC\nBob,25,LA" > "$TEST_DIR/data.csv"
    echo -e "col1\tcol2\tcol3\nval1\tval2\tval3" > "$TEST_DIR/tabs.txt"
    echo -e "  spaces  \n  around  " > "$TEST_DIR/spaces.txt"
    echo "short" > "$TEST_DIR/short.txt"
    echo "This is a very long line that should be folded when using the fold command with a specific width parameter set" > "$TEST_DIR/longline.txt"

    # JSON files
    cat > "$TEST_DIR/test.json" << 'EOF'
{
    "name": "test",
    "version": "1.0.0",
    "items": [1, 2, 3],
    "nested": {
        "key": "value"
    }
}
EOF

    echo '[{"name":"Alice","age":30},{"name":"Bob","age":25}]' > "$TEST_DIR/array.json"

    # YAML files
    cat > "$TEST_DIR/test.yaml" << 'EOF'
name: test
version: 1.0.0
items:
  - 1
  - 2
  - 3
nested:
  key: value
EOF

    # XML files
    cat > "$TEST_DIR/test.xml" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<root>
    <name>test</name>
    <items>
        <item>1</item>
        <item>2</item>
    </items>
</root>
EOF

    # TOML files
    cat > "$TEST_DIR/test.toml" << 'EOF'
name = "test"
version = "1.0.0"

[database]
host = "localhost"
port = 5432
EOF

    # SQL files
    cat > "$TEST_DIR/test.sql" << 'EOF'
SELECT id, name, email FROM users WHERE active = true ORDER BY name;
EOF

    # HTML files
    cat > "$TEST_DIR/test.html" << 'EOF'
<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<h1>Hello World</h1>
<p>This is a test.</p>
</body>
</html>
EOF

    # CSS files
    cat > "$TEST_DIR/test.css" << 'EOF'
body {
    background: #fff;
    color: #333;
}
h1 {
    font-size: 24px;
}
EOF

    # .env file
    cat > "$TEST_DIR/test.env" << 'EOF'
DATABASE_URL=postgres://localhost/test
API_KEY=secret123
DEBUG=true
EOF

    # Binary-ish file
    echo -e "\x00\x01\x02\x03\x04\x05" > "$TEST_DIR/binary.bin"

    # Go source file for generate test
    cat > "$TEST_DIR/sample.go" << 'EOF'
package sample

func Add(a, b int) int {
    return a + b
}

func Subtract(a, b int) int {
    return a - b
}
EOF

    # Create subdirectory structure
    mkdir -p "$TEST_DIR/subdir/nested"
    echo "nested file" > "$TEST_DIR/subdir/nested/file.txt"
    echo "subdir file" > "$TEST_DIR/subdir/file.txt"

    # Proto file
    cat > "$TEST_DIR/test.proto" << 'EOF'
syntax = "proto3";

package test;

message User {
    string name = 1;
    int32 age = 2;
}
EOF

    log_info "Fixtures created"
}

# ============================================================================
# CORE COMMANDS
# ============================================================================
test_core_commands() {
    log_section "CORE COMMANDS"

    # echo
    run_test_exact "echo: basic" "$OMNI echo 'Hello World'" "Hello World"
    run_test_contains "echo: -n (no newline)" "$OMNI echo -n 'test'" "test"
    run_test_contains "echo: -e (escape sequences)" "$OMNI echo -e 'line1\nline2'" "line1"

    # printf
    run_test_contains "printf: basic" "$OMNI printf '%s %d' hello 42" "hello 42"
    run_test_contains "printf: format string" "$OMNI printf '%.2f' 3.14159" "3.14"

    # pwd
    run_test "pwd: basic" "$OMNI pwd"

    # ls
    run_test "ls: current dir" "$OMNI ls"
    run_test "ls: specific dir" "$OMNI ls '$TEST_DIR'"
    run_test_contains "ls: -l (long)" "$OMNI ls -l '$TEST_DIR'" "test.txt"
    run_test "ls: -a (all)" "$OMNI ls -a '$TEST_DIR'"
    run_test "ls: -R (recursive)" "$OMNI ls -R '$TEST_DIR'"
    run_test "ls: -h (human readable)" "$OMNI ls -lh '$TEST_DIR'"
    run_test "ls: --sort=size" "$OMNI ls --sort=size '$TEST_DIR'"
    run_test "ls: --sort=time" "$OMNI ls --sort=time '$TEST_DIR'"

    # cat
    run_test_contains "cat: single file" "$OMNI cat '$TEST_DIR/test.txt'" "Hello World"
    run_test "cat: multiple files" "$OMNI cat '$TEST_DIR/test.txt' '$TEST_DIR/lines.txt'"
    run_test_contains "cat: -n (number lines)" "$OMNI cat -n '$TEST_DIR/test.txt'" "1"
    run_test "cat: -b (number non-blank)" "$OMNI cat -b '$TEST_DIR/test.txt'"
    run_test "cat: -s (squeeze blank)" "$OMNI cat -s '$TEST_DIR/test.txt'"
    run_test "cat: -E (show ends)" "$OMNI cat -E '$TEST_DIR/test.txt'"
    run_test "cat: -T (show tabs)" "$OMNI cat -T '$TEST_DIR/tabs.txt'"

    # tree
    run_test "tree: basic" "$OMNI tree '$TEST_DIR'"
    run_test "tree: -L (level)" "$OMNI tree -L 1 '$TEST_DIR'"
    run_test "tree: -d (dirs only)" "$OMNI tree -d '$TEST_DIR'"
    run_test "tree: -a (all)" "$OMNI tree -a '$TEST_DIR'"
    run_test "tree: --noreport" "$OMNI tree --noreport '$TEST_DIR'"

    # date
    run_test "date: basic" "$OMNI date"
    run_test "date: +format" "$OMNI date '+%Y-%m-%d'"
    run_test "date: -u (UTC)" "$OMNI date -u"
    run_test_contains "date: ISO format" "$OMNI date --iso-8601" "$(date +%Y)"

    # dirname
    run_test_exact "dirname: basic" "$OMNI dirname /path/to/file.txt" "/path/to"
    run_test_exact "dirname: single component" "$OMNI dirname file.txt" "."

    # basename
    run_test_exact "basename: basic" "$OMNI basename /path/to/file.txt" "file.txt"
    run_test_exact "basename: with suffix" "$OMNI basename /path/to/file.txt .txt" "file"

    # realpath
    run_test "realpath: basic" "$OMNI realpath '$TEST_DIR/test.txt'"
    run_test "realpath: -s (no symlink)" "$OMNI realpath -s '$TEST_DIR/test.txt'"

    # readlink
    # Create symlink for testing
    ln -sf "$TEST_DIR/test.txt" "$TEST_DIR/link.txt" 2>/dev/null || true
    run_test "readlink: basic" "$OMNI readlink '$TEST_DIR/link.txt'" || log_skip "readlink: symlinks not supported"

    # arch
    run_test "arch: basic" "$OMNI arch"

    # uname
    run_test "uname: basic" "$OMNI uname"
    run_test "uname: -a (all)" "$OMNI uname -a"
    run_test "uname: -s (kernel)" "$OMNI uname -s"
    run_test "uname: -r (release)" "$OMNI uname -r"
    run_test "uname: -m (machine)" "$OMNI uname -m"

    # sleep
    run_test "sleep: 0.1 seconds" "$OMNI sleep 0.1"

    # seq
    run_test_contains "seq: basic" "$OMNI seq 5" "1"
    run_test_contains "seq: range" "$OMNI seq 2 5" "2"
    run_test_contains "seq: step" "$OMNI seq 1 2 10" "1"
    run_test_contains "seq: -s (separator)" "$OMNI seq -s ',' 3" "1,2,3"
    run_test "seq: -w (equal width)" "$OMNI seq -w 8 12"

    # yes (with timeout)
    run_test "yes: with head" "$OMNI yes | head -n 3"
    run_test "yes: custom string" "$OMNI yes 'custom' | head -n 2"
}

# ============================================================================
# FILE OPERATIONS
# ============================================================================
test_file_commands() {
    log_section "FILE OPERATIONS"

    # touch
    run_test "touch: create file" "$OMNI touch '$TEST_DIR/newfile.txt'"
    run_test "touch: update existing" "$OMNI touch '$TEST_DIR/test.txt'"
    run_test "touch: multiple files" "$OMNI touch '$TEST_DIR/new1.txt' '$TEST_DIR/new2.txt'"

    # mkdir
    run_test "mkdir: basic" "$OMNI mkdir '$TEST_DIR/newdir'"
    run_test "mkdir: -p (parents)" "$OMNI mkdir -p '$TEST_DIR/deep/nested/dir'"
    run_test "mkdir: -m (mode)" "$OMNI mkdir -m 755 '$TEST_DIR/modedir'"

    # cp
    run_test "cp: file" "$OMNI cp '$TEST_DIR/test.txt' '$TEST_DIR/test_copy.txt'"
    run_test "cp: -r (recursive)" "$OMNI cp -r '$TEST_DIR/subdir' '$TEST_DIR/subdir_copy'"
    run_test "cp: -v (verbose)" "$OMNI cp -v '$TEST_DIR/test.txt' '$TEST_DIR/test_copy2.txt'"

    # mv
    run_test "mv: rename" "$OMNI mv '$TEST_DIR/new1.txt' '$TEST_DIR/renamed.txt'"
    run_test "mv: -v (verbose)" "$OMNI mv -v '$TEST_DIR/new2.txt' '$TEST_DIR/renamed2.txt'"

    # rm
    run_test "rm: file" "$OMNI rm '$TEST_DIR/renamed2.txt'"
    run_test "rm: -r (recursive)" "$OMNI rm -r '$TEST_DIR/subdir_copy'"
    run_test "rm: -f (force, nonexistent)" "$OMNI rm -f '$TEST_DIR/nonexistent.txt'"

    # rmdir
    run_test "rmdir: empty dir" "$OMNI rmdir '$TEST_DIR/newdir'"
    run_test "rmdir: -p (parents)" "$OMNI rmdir -p '$TEST_DIR/deep/nested/dir'"

    # stat
    run_test "stat: basic" "$OMNI stat '$TEST_DIR/test.txt'"

    # file
    run_test_contains "file: text" "$OMNI file '$TEST_DIR/test.txt'" "text"
    run_test "file: json" "$OMNI file '$TEST_DIR/test.json'"
    run_test "file: binary" "$OMNI file '$TEST_DIR/binary.bin'"

    # ln (create symlink)
    run_test "ln: symbolic" "$OMNI ln -s '$TEST_DIR/test.txt' '$TEST_DIR/symlink.txt'" || log_skip "ln: symlinks not supported"
    run_test "ln: -f (force)" "$OMNI ln -sf '$TEST_DIR/test.txt' '$TEST_DIR/symlink2.txt'" || log_skip "ln: symlinks not supported"

    # chmod
    run_test "chmod: numeric" "$OMNI chmod 644 '$TEST_DIR/test_copy.txt'"
    run_test "chmod: symbolic" "$OMNI chmod u+x '$TEST_DIR/test_copy.txt'"

    # chown (may need root, test anyway)
    run_test "chown: current user" "$OMNI chown \$(whoami) '$TEST_DIR/test_copy.txt'" || log_skip "chown: permission denied"

    # find
    run_test "find: basic" "$OMNI find '$TEST_DIR'"
    run_test_contains "find: -name" "$OMNI find '$TEST_DIR' -name '*.txt'" "test.txt"
    run_test "find: -type f" "$OMNI find '$TEST_DIR' -type f"
    run_test "find: -type d" "$OMNI find '$TEST_DIR' -type d"
    run_test "find: -maxdepth" "$OMNI find '$TEST_DIR' -maxdepth 1"

    # which
    run_test "which: basic" "$OMNI which ls" || run_test "which: omni" "$OMNI which omni"

    # dd (careful with this one)
    run_test "dd: basic copy" "$OMNI dd if='$TEST_DIR/test.txt' of='$TEST_DIR/dd_out.txt'"
    run_test "dd: count" "$OMNI dd if='$TEST_DIR/test.txt' of='$TEST_DIR/dd_out2.txt' bs=1 count=10"
}

# ============================================================================
# TEXT PROCESSING
# ============================================================================
test_text_commands() {
    log_section "TEXT PROCESSING"

    # head
    run_test "head: basic" "$OMNI head '$TEST_DIR/lines.txt'"
    run_test_contains "head: -n 2" "$OMNI head -n 2 '$TEST_DIR/lines.txt'" "line1"
    run_test "head: -c (bytes)" "$OMNI head -c 10 '$TEST_DIR/test.txt'"

    # tail
    run_test "tail: basic" "$OMNI tail '$TEST_DIR/lines.txt'"
    run_test_contains "tail: -n 2" "$OMNI tail -n 2 '$TEST_DIR/lines.txt'" "line5"
    run_test "tail: -c (bytes)" "$OMNI tail -c 10 '$TEST_DIR/test.txt'"

    # wc
    run_test "wc: basic" "$OMNI wc '$TEST_DIR/test.txt'"
    run_test "wc: -l (lines)" "$OMNI wc -l '$TEST_DIR/lines.txt'"
    run_test "wc: -w (words)" "$OMNI wc -w '$TEST_DIR/test.txt'"
    run_test "wc: -c (bytes)" "$OMNI wc -c '$TEST_DIR/test.txt'"
    run_test "wc: -m (chars)" "$OMNI wc -m '$TEST_DIR/test.txt'"

    # sort
    run_test "sort: basic" "$OMNI sort '$TEST_DIR/duplicates.txt'"
    run_test "sort: -r (reverse)" "$OMNI sort -r '$TEST_DIR/duplicates.txt'"
    run_test "sort: -n (numeric)" "$OMNI sort -n '$TEST_DIR/numbers.txt'"
    run_test "sort: -u (unique)" "$OMNI sort -u '$TEST_DIR/duplicates.txt'"
    run_test "sort: -k (key)" "$OMNI sort -t',' -k2 '$TEST_DIR/data.csv'"

    # uniq
    run_test "uniq: basic (presorted)" "echo -e 'a\na\nb\nb\nc' | $OMNI uniq"
    run_test "uniq: -c (count)" "echo -e 'a\na\nb\nc' | $OMNI uniq -c"
    run_test "uniq: -d (duplicates only)" "echo -e 'a\na\nb\nc' | $OMNI uniq -d"
    run_test "uniq: -u (unique only)" "echo -e 'a\na\nb\nc' | $OMNI uniq -u"

    # grep
    run_test_contains "grep: basic" "$OMNI grep 'Hello' '$TEST_DIR/test.txt'" "Hello"
    run_test "grep: -i (ignore case)" "$OMNI grep -i 'hello' '$TEST_DIR/test.txt'"
    run_test "grep: -v (invert)" "$OMNI grep -v 'Hello' '$TEST_DIR/test.txt'"
    run_test "grep: -n (line numbers)" "$OMNI grep -n 'line' '$TEST_DIR/lines.txt'"
    run_test "grep: -c (count)" "$OMNI grep -c 'line' '$TEST_DIR/lines.txt'"
    run_test "grep: -l (files only)" "$OMNI grep -l 'Hello' '$TEST_DIR'/*.txt"
    run_test "grep: -r (recursive)" "$OMNI grep -r 'file' '$TEST_DIR/subdir'"
    run_test "grep: -E (extended regex)" "$OMNI grep -E 'line[0-9]' '$TEST_DIR/lines.txt'"
    run_test "grep: -w (word)" "$OMNI grep -w 'Hello' '$TEST_DIR/test.txt'"

    # egrep (alias)
    run_test "egrep: extended regex" "$OMNI egrep 'line[0-9]' '$TEST_DIR/lines.txt'"

    # fgrep (alias)
    run_test "fgrep: fixed string" "$OMNI fgrep 'Hello' '$TEST_DIR/test.txt'"

    # cut
    run_test "cut: -d -f (delimiter, field)" "$OMNI cut -d',' -f1 '$TEST_DIR/data.csv'"
    run_test "cut: -c (chars)" "$OMNI cut -c1-5 '$TEST_DIR/test.txt'"
    run_test "cut: multiple fields" "$OMNI cut -d',' -f1,3 '$TEST_DIR/data.csv'"

    # tr
    run_test "tr: basic" "echo 'hello' | $OMNI tr 'a-z' 'A-Z'"
    run_test "tr: -d (delete)" "echo 'hello123' | $OMNI tr -d '0-9'"
    run_test "tr: -s (squeeze)" "echo 'heeello' | $OMNI tr -s 'e'"
    run_test "tr: -c (complement)" "echo 'hello123' | $OMNI tr -cd '0-9'"

    # sed
    run_test "sed: substitute" "echo 'hello world' | $OMNI sed 's/world/universe/'"
    run_test "sed: global" "echo 'aaa' | $OMNI sed 's/a/b/g'"
    run_test "sed: delete line" "$OMNI sed '1d' '$TEST_DIR/lines.txt'"
    run_test "sed: print line" "$OMNI sed -n '2p' '$TEST_DIR/lines.txt'"
    run_test "sed: address range" "$OMNI sed '2,4d' '$TEST_DIR/lines.txt'"

    # awk
    run_test "awk: print field" "echo 'a b c' | $OMNI awk '{print \$2}'"
    run_test "awk: -F (field separator)" "$OMNI awk -F',' '{print \$1}' '$TEST_DIR/data.csv'"
    run_test "awk: NR (line number)" "$OMNI awk '{print NR, \$0}' '$TEST_DIR/lines.txt'"
    run_test "awk: pattern" "$OMNI awk '/line2/' '$TEST_DIR/lines.txt'"
    run_test "awk: BEGIN/END" "echo '1\n2\n3' | $OMNI awk 'BEGIN{s=0}{s+=\$1}END{print s}'"

    # nl
    run_test "nl: basic" "$OMNI nl '$TEST_DIR/lines.txt'"
    run_test "nl: -b a (all lines)" "$OMNI nl -b a '$TEST_DIR/test.txt'"
    run_test "nl: -n (format)" "$OMNI nl -n rz '$TEST_DIR/lines.txt'"

    # paste
    run_test "paste: files" "$OMNI paste '$TEST_DIR/short.txt' '$TEST_DIR/short.txt'"
    run_test "paste: -d (delimiter)" "$OMNI paste -d',' '$TEST_DIR/short.txt' '$TEST_DIR/short.txt'"
    run_test "paste: -s (serial)" "$OMNI paste -s '$TEST_DIR/lines.txt'"

    # tac
    run_test_contains "tac: reverse lines" "$OMNI tac '$TEST_DIR/lines.txt'" "line5"

    # rev
    run_test "rev: reverse chars" "echo 'hello' | $OMNI rev"

    # column
    run_test "column: basic" "$OMNI column '$TEST_DIR/data.csv'"
    run_test "column: -t (table)" "$OMNI column -t '$TEST_DIR/tabs.txt'"
    run_test "column: -s (separator)" "$OMNI column -t -s',' '$TEST_DIR/data.csv'"

    # fold
    run_test "fold: basic" "$OMNI fold '$TEST_DIR/longline.txt'"
    run_test "fold: -w (width)" "$OMNI fold -w 20 '$TEST_DIR/longline.txt'"
    run_test "fold: -s (spaces)" "$OMNI fold -s -w 20 '$TEST_DIR/longline.txt'"

    # join
    echo -e "1 Alice\n2 Bob" > "$TEST_DIR/join1.txt"
    echo -e "1 NYC\n2 LA" > "$TEST_DIR/join2.txt"
    run_test "join: basic" "$OMNI join '$TEST_DIR/join1.txt' '$TEST_DIR/join2.txt'"

    # comm
    echo -e "a\nb\nc" | sort > "$TEST_DIR/comm1.txt"
    echo -e "b\nc\nd" | sort > "$TEST_DIR/comm2.txt"
    run_test "comm: basic" "$OMNI comm '$TEST_DIR/comm1.txt' '$TEST_DIR/comm2.txt'"
    run_test "comm: -1 (suppress col 1)" "$OMNI comm -1 '$TEST_DIR/comm1.txt' '$TEST_DIR/comm2.txt'"
    run_test "comm: -12 (common only)" "$OMNI comm -12 '$TEST_DIR/comm1.txt' '$TEST_DIR/comm2.txt'"

    # cmp
    run_test "cmp: identical" "$OMNI cmp '$TEST_DIR/test.txt' '$TEST_DIR/test.txt'"
    run_test "cmp: different" "$OMNI cmp '$TEST_DIR/test.txt' '$TEST_DIR/lines.txt'" 1
    run_test "cmp: -s (silent)" "$OMNI cmp -s '$TEST_DIR/test.txt' '$TEST_DIR/test.txt'"

    # diff
    run_test "diff: identical" "$OMNI diff '$TEST_DIR/test.txt' '$TEST_DIR/test.txt'"
    run_test "diff: different" "$OMNI diff '$TEST_DIR/test.txt' '$TEST_DIR/lines.txt'" || true
    run_test "diff: -u (unified)" "$OMNI diff -u '$TEST_DIR/test.txt' '$TEST_DIR/lines.txt'" || true

    # strings
    run_test "strings: basic" "$OMNI strings '$TEST_DIR/test.txt'"
    run_test "strings: -n (min length)" "$OMNI strings -n 5 '$TEST_DIR/test.txt'"

    # shuf
    run_test "shuf: basic" "$OMNI shuf '$TEST_DIR/lines.txt'"
    run_test "shuf: -n (count)" "$OMNI shuf -n 2 '$TEST_DIR/lines.txt'"
    run_test "shuf: -e (args)" "$OMNI shuf -e a b c d"

    # split
    run_test "split: basic" "$OMNI split -l 2 '$TEST_DIR/lines.txt' '$TEST_DIR/split_'"
    run_test "split: -b (bytes)" "$OMNI split -b 10 '$TEST_DIR/test.txt' '$TEST_DIR/split_b_'"
}

# ============================================================================
# SYSTEM COMMANDS
# ============================================================================
test_system_commands() {
    log_section "SYSTEM COMMANDS"

    # env
    run_test "env: basic" "$OMNI env"
    run_test "env: specific var" "$OMNI env | grep -q PATH || true"

    # whoami
    run_test "whoami: basic" "$OMNI whoami"

    # id
    run_test "id: basic" "$OMNI id"
    run_test "id: -u (uid)" "$OMNI id -u"
    run_test "id: -g (gid)" "$OMNI id -g"
    run_test "id: -n (name)" "$OMNI id -un"

    # uptime
    run_test "uptime: basic" "$OMNI uptime"

    # free (may not work on all systems)
    run_test "free: basic" "$OMNI free" || log_skip "free: not available"
    run_test "free: -h (human)" "$OMNI free -h" || log_skip "free: not available"
    run_test "free: -m (MB)" "$OMNI free -m" || log_skip "free: not available"

    # df
    run_test "df: basic" "$OMNI df"
    run_test "df: -h (human)" "$OMNI df -h"
    run_test "df: specific path" "$OMNI df '$TEST_DIR'"

    # du
    run_test "du: basic" "$OMNI du '$TEST_DIR'"
    run_test "du: -h (human)" "$OMNI du -h '$TEST_DIR'"
    run_test "du: -s (summary)" "$OMNI du -s '$TEST_DIR'"
    run_test "du: -d (depth)" "$OMNI du -d 1 '$TEST_DIR'"

    # ps
    run_test "ps: basic" "$OMNI ps"
    run_test "ps: -a (all)" "$OMNI ps -a" || true
    run_test "ps: -e (every)" "$OMNI ps -e" || true

    # kill (just test --list)
    run_test "kill: -l (list signals)" "$OMNI kill -l"

    # time
    run_test "time: basic" "$OMNI time $OMNI echo test"
}

# ============================================================================
# ARCHIVE & COMPRESSION
# ============================================================================
test_archive_commands() {
    log_section "ARCHIVE & COMPRESSION"

    # Create test content
    echo "archive test content" > "$TEST_DIR/archive_test.txt"
    mkdir -p "$TEST_DIR/archive_dir"
    echo "file1" > "$TEST_DIR/archive_dir/file1.txt"
    echo "file2" > "$TEST_DIR/archive_dir/file2.txt"

    # tar
    run_test "tar: create" "$OMNI tar -cf '$TEST_DIR/test.tar' -C '$TEST_DIR' archive_test.txt"
    run_test "tar: list" "$OMNI tar -tf '$TEST_DIR/test.tar'"
    run_test "tar: extract" "$OMNI tar -xf '$TEST_DIR/test.tar' -C '$TEST_DIR/archive_dir'"
    run_test "tar: create gzip" "$OMNI tar -czf '$TEST_DIR/test.tar.gz' -C '$TEST_DIR' archive_test.txt"
    run_test "tar: list gzip" "$OMNI tar -tzf '$TEST_DIR/test.tar.gz'"
    run_test "tar: extract gzip" "$OMNI tar -xzf '$TEST_DIR/test.tar.gz' -C '$TEST_DIR/archive_dir'"

    # zip
    run_test "zip: create" "$OMNI zip '$TEST_DIR/test.zip' '$TEST_DIR/archive_test.txt'"
    run_test "zip: add file" "$OMNI zip '$TEST_DIR/test.zip' '$TEST_DIR/test.txt'"

    # unzip
    run_test "unzip: list" "$OMNI unzip -l '$TEST_DIR/test.zip'"
    run_test "unzip: extract" "$OMNI unzip -d '$TEST_DIR/unzip_out' '$TEST_DIR/test.zip'"

    # gzip
    cp "$TEST_DIR/archive_test.txt" "$TEST_DIR/gzip_test.txt"
    run_test "gzip: compress" "$OMNI gzip '$TEST_DIR/gzip_test.txt'"

    # gunzip
    run_test "gunzip: decompress" "$OMNI gunzip '$TEST_DIR/gzip_test.txt.gz'"

    # zcat
    cp "$TEST_DIR/archive_test.txt" "$TEST_DIR/zcat_test.txt"
    $OMNI gzip "$TEST_DIR/zcat_test.txt"
    run_test "zcat: read" "$OMNI zcat '$TEST_DIR/zcat_test.txt.gz'"

    # bzip2
    cp "$TEST_DIR/archive_test.txt" "$TEST_DIR/bzip_test.txt"
    run_test "bzip2: compress" "$OMNI bzip2 '$TEST_DIR/bzip_test.txt'"

    # bunzip2
    run_test "bunzip2: decompress" "$OMNI bunzip2 '$TEST_DIR/bzip_test.txt.bz2'"

    # bzcat
    cp "$TEST_DIR/archive_test.txt" "$TEST_DIR/bzcat_test.txt"
    $OMNI bzip2 "$TEST_DIR/bzcat_test.txt"
    run_test "bzcat: read" "$OMNI bzcat '$TEST_DIR/bzcat_test.txt.bz2'"

    # xz
    cp "$TEST_DIR/archive_test.txt" "$TEST_DIR/xz_test.txt"
    run_test "xz: compress" "$OMNI xz '$TEST_DIR/xz_test.txt'"

    # unxz
    run_test "unxz: decompress" "$OMNI unxz '$TEST_DIR/xz_test.txt.xz'"

    # xzcat
    cp "$TEST_DIR/archive_test.txt" "$TEST_DIR/xzcat_test.txt"
    $OMNI xz "$TEST_DIR/xzcat_test.txt"
    run_test "xzcat: read" "$OMNI xzcat '$TEST_DIR/xzcat_test.txt.xz'"
}

# ============================================================================
# HASH & ENCODING
# ============================================================================
test_hash_commands() {
    log_section "HASH & ENCODING"

    # hash
    run_test "hash: sha256" "echo -n 'test' | $OMNI hash sha256"
    run_test "hash: sha512" "echo -n 'test' | $OMNI hash sha512"
    run_test "hash: md5" "echo -n 'test' | $OMNI hash md5"
    run_test "hash: file" "$OMNI hash sha256 '$TEST_DIR/test.txt'"

    # sha256sum
    run_test "sha256sum: basic" "$OMNI sha256sum '$TEST_DIR/test.txt'"
    run_test "sha256sum: stdin" "echo -n 'test' | $OMNI sha256sum"

    # sha512sum
    run_test "sha512sum: basic" "$OMNI sha512sum '$TEST_DIR/test.txt'"

    # md5sum
    run_test "md5sum: basic" "$OMNI md5sum '$TEST_DIR/test.txt'"
}

test_encoding_commands() {
    log_section "ENCODING"

    # base64
    run_test "base64: encode" "echo -n 'hello' | $OMNI base64"
    run_test "base64: decode" "echo -n 'aGVsbG8=' | $OMNI base64 -d"
    run_test "base64: file" "$OMNI base64 '$TEST_DIR/test.txt'"

    # base32
    run_test "base32: encode" "echo -n 'hello' | $OMNI base32"
    run_test "base32: decode" "echo -n 'NBSWY3DP' | $OMNI base32 -d"

    # base58
    run_test "base58: encode" "echo -n 'hello' | $OMNI base58"
    run_test "base58: decode" "$OMNI base58 -d" <<< "Cn8eVZg"

    # hex
    run_test "hex: encode" "echo -n 'hello' | $OMNI hex encode"
    run_test "hex: decode" "echo -n '68656c6c6f' | $OMNI hex decode"

    # url
    run_test "url: encode" "$OMNI url encode 'hello world'"
    run_test "url: decode" "$OMNI url decode 'hello%20world'"

    # html (encode/decode)
    run_test "html: encode" "$OMNI html encode '<div>test</div>'"
    run_test "html: decode" "$OMNI html decode '&lt;div&gt;test&lt;/div&gt;'"

    # xxd
    run_test_contains "xxd: basic dump" "echo -n 'hello' | $OMNI xxd" "68656c6c6f"
    run_test_contains "xxd: with address" "echo -n 'hello' | $OMNI xxd" "00000000:"
    run_test "xxd: plain mode" "echo -n 'hello' | $OMNI xxd -p"
    run_test "xxd: include mode" "echo -n 'Hi' | $OMNI xxd -i"
    run_test "xxd: binary mode" "echo -n 'A' | $OMNI xxd -b"
    run_test "xxd: uppercase" "echo -n 'hello' | $OMNI xxd -u"
    run_test "xxd: custom columns" "echo -n 'hello world' | $OMNI xxd -c 8"
    run_test "xxd: single byte groups" "echo -n 'hello' | $OMNI xxd -g 1"
    run_test "xxd: reverse plain" "echo '68656c6c6f' | $OMNI xxd -r -p"
    run_test "xxd: roundtrip" "echo -n 'test' | $OMNI xxd -p | $OMNI xxd -r -p"
    run_test "xxd: length limit" "echo -n 'hello world' | $OMNI xxd -l 5"
    run_test "xxd: file" "$OMNI xxd '$TEST_DIR/test.txt'"
}

# ============================================================================
# DATA PROCESSING
# ============================================================================
test_data_commands() {
    log_section "DATA PROCESSING"

    # jq
    run_test_contains "jq: query" "$OMNI jq '.name' '$TEST_DIR/test.json'" "test"
    run_test "jq: array" "$OMNI jq '.items[0]' '$TEST_DIR/test.json'"
    run_test "jq: nested" "$OMNI jq '.nested.key' '$TEST_DIR/test.json'"
    run_test "jq: stdin" "echo '{\"a\":1}' | $OMNI jq '.a'"

    # yq
    run_test_contains "yq: query" "$OMNI yq '.name' '$TEST_DIR/test.yaml'" "test"
    run_test "yq: array" "$OMNI yq '.items[0]' '$TEST_DIR/test.yaml'"
    run_test "yq: nested" "$OMNI yq '.nested.key' '$TEST_DIR/test.yaml'"

    # json fmt
    run_test "json: fmt" "$OMNI json fmt '$TEST_DIR/test.json'"
    run_test "json: fmt (stdin)" "echo '{\"a\":1}' | $OMNI json fmt"
    run_test "json: fmt --indent" "$OMNI json fmt --indent 4 '$TEST_DIR/test.json'"

    # json minify
    run_test "json: minify" "$OMNI json minify '$TEST_DIR/test.json'"

    # json validate
    run_test "json: validate" "$OMNI json validate '$TEST_DIR/test.json'"
    run_test "json: validate invalid" "echo '{invalid}' | $OMNI json validate" 1

    # json stats
    run_test "json: stats" "$OMNI json stats '$TEST_DIR/test.json'"

    # json keys
    run_test "json: keys" "$OMNI json keys '$TEST_DIR/test.json'"

    # json tostruct
    run_test "json: tostruct" "$OMNI json tostruct '$TEST_DIR/test.json'"

    # json tocsv
    run_test "json: tocsv" "$OMNI json tocsv '$TEST_DIR/array.json'"

    # json fromcsv
    run_test "json: fromcsv" "$OMNI json fromcsv '$TEST_DIR/data.csv'"

    # json toyaml
    run_test "json: toyaml" "$OMNI json toyaml '$TEST_DIR/test.json'"

    # json toxml
    run_test "json: toxml" "$OMNI json toxml '$TEST_DIR/test.json'"

    # yaml validate
    run_test "yaml: validate" "$OMNI yaml validate '$TEST_DIR/test.yaml'"

    # yaml fmt
    run_test "yaml: fmt" "$OMNI yaml fmt '$TEST_DIR/test.yaml'"

    # yaml tostruct
    run_test "yaml: tostruct" "$OMNI yaml tostruct '$TEST_DIR/test.yaml'"

    # xml fmt
    run_test "xml: fmt" "$OMNI xml fmt '$TEST_DIR/test.xml'"

    # xml validate
    run_test "xml: validate" "$OMNI xml validate '$TEST_DIR/test.xml'"

    # xml tojson
    run_test "xml: tojson" "$OMNI xml tojson '$TEST_DIR/test.xml'"

    # toml validate
    run_test "toml: validate" "$OMNI toml validate '$TEST_DIR/test.toml'"

    # toml fmt
    run_test "toml: fmt" "$OMNI toml fmt '$TEST_DIR/test.toml'"

    # sql fmt
    run_test "sql: fmt" "$OMNI sql fmt '$TEST_DIR/test.sql'"
    run_test "sql: fmt stdin" "echo 'select a,b from t where x=1' | $OMNI sql fmt"

    # sql minify
    run_test "sql: minify" "$OMNI sql minify '$TEST_DIR/test.sql'"

    # sql validate
    run_test "sql: validate" "$OMNI sql validate '$TEST_DIR/test.sql'"

    # html fmt
    run_test "html: fmt" "$OMNI html fmt '$TEST_DIR/test.html'"

    # html minify
    run_test "html: minify" "$OMNI html minify '$TEST_DIR/test.html'"

    # html validate
    run_test "html: validate" "$OMNI html validate '$TEST_DIR/test.html'"

    # css fmt
    run_test "css: fmt" "$OMNI css fmt '$TEST_DIR/test.css'"

    # css minify
    run_test "css: minify" "$OMNI css minify '$TEST_DIR/test.css'"

    # css validate
    run_test "css: validate" "$OMNI css validate '$TEST_DIR/test.css'"

    # dotenv
    run_test "dotenv: basic" "$OMNI dotenv '$TEST_DIR/test.env'"
    run_test_contains "dotenv: get" "$OMNI dotenv '$TEST_DIR/test.env' API_KEY" "secret123"

    # csv
    run_test "csv: basic" "$OMNI csv '$TEST_DIR/data.csv'" || true
}

# ============================================================================
# ID GENERATORS
# ============================================================================
test_id_generators() {
    log_section "ID GENERATORS"

    # uuid
    run_test "uuid: v4" "$OMNI uuid"
    run_test "uuid: v1" "$OMNI uuid v1"
    run_test "uuid: v4 explicit" "$OMNI uuid v4"
    run_test "uuid: v7" "$OMNI uuid v7"

    # random
    run_test "random: string" "$OMNI random string 16"
    run_test "random: hex" "$OMNI random hex 16"
    run_test "random: int" "$OMNI random int 1 100"
    run_test "random: password" "$OMNI random password 16"
    run_test "random: bytes" "$OMNI random bytes 16"

    # ksuid
    run_test "ksuid: generate" "$OMNI ksuid"

    # ulid
    run_test "ulid: generate" "$OMNI ulid"

    # snowflake
    run_test "snowflake: generate" "$OMNI snowflake"

    # nanoid
    run_test "nanoid: generate" "$OMNI nanoid"
    run_test "nanoid: custom length" "$OMNI nanoid -l 32"
}

# ============================================================================
# SECURITY & CRYPTO
# ============================================================================
test_security_commands() {
    log_section "SECURITY & CRYPTO"

    # encrypt/decrypt
    run_test "encrypt: basic" "echo 'secret data' | $OMNI encrypt -p 'password123' > '$TEST_DIR/encrypted.bin'"
    run_test "decrypt: basic" "$OMNI decrypt -p 'password123' '$TEST_DIR/encrypted.bin'"

    # jwt
    JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
    run_test "jwt: decode" "$OMNI jwt decode '$JWT_TOKEN'"
}

# ============================================================================
# UTILITIES
# ============================================================================
test_utility_commands() {
    log_section "UTILITIES"

    # xargs
    run_test "xargs: basic" "echo 'a b c' | $OMNI xargs echo"
    run_test "xargs: -n" "echo 'a b c d' | $OMNI xargs -n 2 echo"
    run_test "xargs: -I" "echo 'a\nb' | $OMNI xargs -I {} echo 'item: {}'"

    # watch (quick test)
    run_test "watch: single run" "timeout 1 $OMNI watch -n 0.1 'echo test'" || true

    # for
    run_test "for: basic" "$OMNI for i in 1 2 3 -- echo 'num: \$i'"

    # case
    run_test "case: upper" "$OMNI case upper 'hello world'"
    run_test "case: lower" "$OMNI case lower 'HELLO WORLD'"
    run_test "case: title" "$OMNI case title 'hello world'"
    run_test "case: snake" "$OMNI case snake 'helloWorld'"
    run_test "case: camel" "$OMNI case camel 'hello_world'"
    run_test "case: kebab" "$OMNI case kebab 'hello_world'"

    # cron
    run_test "cron: parse" "$OMNI cron parse '0 * * * *'"
    run_test "cron: next" "$OMNI cron next '0 * * * *'"

    # loc
    run_test "loc: basic" "$OMNI loc '$TEST_DIR'"
    run_test "loc: specific" "$OMNI loc '$TEST_DIR/sample.go'"

    # cmdtree
    run_test "cmdtree: basic" "$OMNI cmdtree"

    # aicontext
    run_test "aicontext: basic" "$OMNI aicontext" || log_skip "aicontext: not available"

    # lint (on the test Go file)
    run_test "lint: go file" "$OMNI lint '$TEST_DIR/sample.go'" || log_skip "lint: not available"
}

# ============================================================================
# CODE GENERATION
# ============================================================================
test_generate_commands() {
    log_section "CODE GENERATION"

    # generate test
    run_test "generate: test" "$OMNI generate test '$TEST_DIR/sample.go'" || log_skip "generate test: not available"

    # generate handler
    run_test "generate: handler" "$OMNI generate handler User" || log_skip "generate handler: not available"

    # generate repository
    run_test "generate: repository" "$OMNI generate repository User" || log_skip "generate repository: not available"
}

# ============================================================================
# PROTOBUF (buf)
# ============================================================================
test_buf_commands() {
    log_section "PROTOBUF (buf)"

    # buf lint
    run_test "buf: lint" "$OMNI buf lint '$TEST_DIR/test.proto'" || log_skip "buf lint: not available"

    # buf format
    run_test "buf: format" "$OMNI buf format '$TEST_DIR/test.proto'" || log_skip "buf format: not available"

    # buf ls-files
    run_test "buf: ls-files" "$OMNI buf ls-files '$TEST_DIR'" || log_skip "buf ls-files: not available"
}

# ============================================================================
# DATABASE
# ============================================================================
test_db_commands() {
    log_section "DATABASE"

    # sqlite
    run_test "sqlite: create and query" "$OMNI sqlite '$TEST_DIR/test.db' 'CREATE TABLE t(id INT); INSERT INTO t VALUES(1); SELECT * FROM t;'" || log_skip "sqlite: not available"

    # bbolt
    run_test "bbolt: create" "$OMNI bbolt create '$TEST_DIR/test.bolt'" || log_skip "bbolt: not available"
}

# ============================================================================
# PAGERS (interactive - just verify they exist)
# ============================================================================
test_pager_commands() {
    log_section "PAGERS"

    # less (non-interactive test)
    run_test "less: --version or help" "$OMNI less --help" || log_skip "less: not available"

    # more (non-interactive test)
    run_test "more: --version or help" "$OMNI more --help" || log_skip "more: not available"
}

# ============================================================================
# SPECIAL
# ============================================================================
test_special_commands() {
    log_section "SPECIAL"

    # pipe
    run_test "pipe: basic" "echo 'hello' | $OMNI pipe '$OMNI tr a-z A-Z'"

    # gops
    run_test "gops: basic" "$OMNI gops" || log_skip "gops: not available"

    # top (non-interactive)
    run_test "top: -b (batch)" "timeout 1 $OMNI top -b -n 1" || log_skip "top: not available"

    # curl
    run_test "curl: basic" "$OMNI curl --help" || log_skip "curl: not available"
}

# ============================================================================
# STDIN/PIPE TESTS
# ============================================================================
test_stdin_features() {
    log_section "STDIN/PIPE FEATURES"

    # Test stdin reading for various commands
    run_test "stdin: cat" "echo 'hello' | $OMNI cat"
    run_test "stdin: wc" "echo 'hello world' | $OMNI wc -w"
    run_test "stdin: sort" "echo -e 'c\na\nb' | $OMNI sort"
    run_test "stdin: uniq" "echo -e 'a\na\nb' | $OMNI uniq"
    run_test "stdin: head" "echo -e 'a\nb\nc\nd\ne' | $OMNI head -n 2"
    run_test "stdin: tail" "echo -e 'a\nb\nc\nd\ne' | $OMNI tail -n 2"
    run_test "stdin: grep" "echo -e 'hello\nworld' | $OMNI grep 'world'"
    run_test "stdin: cut" "echo 'a,b,c' | $OMNI cut -d',' -f2"
    run_test "stdin: tr" "echo 'hello' | $OMNI tr 'a-z' 'A-Z'"
    run_test "stdin: sed" "echo 'hello' | $OMNI sed 's/hello/world/'"
    run_test "stdin: awk" "echo 'a b c' | $OMNI awk '{print \$2}'"
    run_test "stdin: base64" "echo -n 'hello' | $OMNI base64"
    run_test "stdin: hash" "echo -n 'test' | $OMNI hash sha256"
    run_test "stdin: json fmt" "echo '{\"a\":1}' | $OMNI json fmt"

    # Pipeline tests
    run_test "pipeline: grep | wc" "echo -e 'a\nb\na\nc' | $OMNI grep 'a' | $OMNI wc -l"
    run_test "pipeline: sort | uniq" "echo -e 'b\na\na\nc' | $OMNI sort | $OMNI uniq"
    run_test "pipeline: cat | tr | sort" "$OMNI cat '$TEST_DIR/lines.txt' | $OMNI tr 'a-z' 'A-Z' | $OMNI sort"
}

# ============================================================================
# ERROR HANDLING TESTS
# ============================================================================
test_error_handling() {
    log_section "ERROR HANDLING"

    # Test error cases
    run_test "error: nonexistent file" "$OMNI cat '$TEST_DIR/nonexistent.txt'" 1
    run_test "error: invalid json" "echo 'not json' | $OMNI json validate" 1
    run_test "error: invalid yaml" "echo ':: invalid' | $OMNI yaml validate" 1
    run_test "error: grep no match" "$OMNI grep 'xyz123' '$TEST_DIR/test.txt'" 1
    run_test "error: invalid base64" "echo 'not-base64!' | $OMNI base64 -d" 1
}

# ============================================================================
# MAIN
# ============================================================================

main() {
    echo ""
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║                    OMNI - Comprehensive Command Test Suite                    ║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    # Check if omni exists
    if [ ! -f "$OMNI" ]; then
        log_info "Building omni..."
        go build -o "$OMNI" . || { echo "Failed to build omni"; exit 1; }
    fi

    log_info "Using omni binary: $OMNI"
    log_info "Test directory: $TEST_DIR"
    log_info "Category: $CATEGORY"
    log_info "Verbose: $VERBOSE"
    echo ""

    create_fixtures

    case "$CATEGORY" in
        "core")
            test_core_commands
            ;;
        "file")
            test_file_commands
            ;;
        "text")
            test_text_commands
            ;;
        "system")
            test_system_commands
            ;;
        "archive")
            test_archive_commands
            ;;
        "hash")
            test_hash_commands
            ;;
        "encoding")
            test_encoding_commands
            ;;
        "data")
            test_data_commands
            ;;
        "util")
            test_utility_commands
            ;;
        "id")
            test_id_generators
            ;;
        "security")
            test_security_commands
            ;;
        "generate")
            test_generate_commands
            ;;
        "buf")
            test_buf_commands
            ;;
        "db")
            test_db_commands
            ;;
        "pager")
            test_pager_commands
            ;;
        "special")
            test_special_commands
            ;;
        "stdin")
            test_stdin_features
            ;;
        "error")
            test_error_handling
            ;;
        "all")
            test_core_commands
            test_file_commands
            test_text_commands
            test_system_commands
            test_archive_commands
            test_hash_commands
            test_encoding_commands
            test_data_commands
            test_id_generators
            test_security_commands
            test_utility_commands
            test_generate_commands
            test_buf_commands
            test_db_commands
            test_pager_commands
            test_special_commands
            test_stdin_features
            test_error_handling
            ;;
        *)
            echo "Unknown category: $CATEGORY"
            echo "Valid categories: core, file, text, system, archive, hash, encoding, data, util, id, security, generate, buf, db, pager, special, stdin, error, all"
            exit 1
            ;;
    esac

    # Summary
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  TEST SUMMARY${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  Total:   ${TOTAL}"
    echo -e "  ${GREEN}Passed:  ${PASSED}${NC}"
    echo -e "  ${RED}Failed:  ${FAILED}${NC}"
    echo -e "  ${YELLOW}Skipped: ${SKIPPED}${NC}"
    echo ""

    if [ "$FAILED" -gt 0 ]; then
        echo -e "${RED}Some tests failed!${NC}"
        exit 1
    else
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    fi
}

main "$@"

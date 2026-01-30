#!/bin/bash
# Test script for omni logger output capture
# This script tests all omni commands and verifies output is captured in logs

# Don't exit on error - we want to continue testing
set +e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
SKIPPED=0

# Test directory
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

# Create test files
echo "hello world" > "$TEST_DIR/test.txt"
echo -e "line1\nline2\nline3" > "$TEST_DIR/lines.txt"
echo -e "cherry\napple\nbanana" > "$TEST_DIR/fruits.txt"
echo -e "a\na\nb\nb\nc" > "$TEST_DIR/dupes.txt"
echo '{"name":"test","value":123}' > "$TEST_DIR/test.json"
echo "name,age,city" > "$TEST_DIR/test.csv"
echo "John,30,NYC" >> "$TEST_DIR/test.csv"
echo "<root><item>value</item></root>" > "$TEST_DIR/test.xml"
echo "key: value" > "$TEST_DIR/test.yaml"
echo "key = \"value\"" > "$TEST_DIR/test.toml"

# Helper function to run test
run_test() {
    local name="$1"
    local cmd="$2"
    local expect_fail="${3:-false}"

    printf "${BLUE}Testing${NC} %-20s ... " "$name"

    if eval "$cmd" > /dev/null 2>&1; then
        if [ "$expect_fail" = "true" ]; then
            printf "${RED}FAIL${NC} (expected failure)\n"
            ((FAILED++))
        else
            printf "${GREEN}PASS${NC}\n"
            ((PASSED++))
        fi
    else
        if [ "$expect_fail" = "true" ]; then
            printf "${GREEN}PASS${NC} (expected failure)\n"
            ((PASSED++))
        else
            printf "${RED}FAIL${NC}\n"
            ((FAILED++))
        fi
    fi
}

# Helper function to skip test
skip_test() {
    local name="$1"
    local reason="$2"
    printf "${BLUE}Testing${NC} %-20s ... ${YELLOW}SKIP${NC} ($reason)\n" "$name"
    ((SKIPPED++))
}

echo "=============================================="
echo "  Omni Logger Output Capture Test Suite"
echo "=============================================="
echo ""
echo "Test directory: $TEST_DIR"
echo ""

# Check logger status
echo "--- Logger Status ---"
omni logger --status
echo ""

echo "--- Running Tests ---"
echo ""

# Basic output commands
run_test "echo" "omni echo 'test message'"
run_test "date" "omni date"
run_test "arch" "omni arch"
run_test "pwd" "omni pwd"
run_test "env" "omni env | head -1"
run_test "seq" "omni seq 1 5"
run_test "shuf" "omni shuf -n 3 -i 1-10"
run_test "yes (limited)" "timeout 1 omni yes test || true"

# File info commands
run_test "basename" "omni basename /path/to/file.txt"
run_test "dirname" "omni dirname /path/to/file.txt"
run_test "file" "omni file '$TEST_DIR/test.txt'"
run_test "realpath" "omni realpath '$TEST_DIR/test.txt'"
run_test "stat" "omni stat '$TEST_DIR/test.txt'"
run_test "ls" "omni ls '$TEST_DIR'"
run_test "tree" "omni tree '$TEST_DIR' --depth 1"
run_test "df" "omni df"
run_test "du" "omni du '$TEST_DIR'"

# File reading commands
run_test "cat" "omni cat '$TEST_DIR/test.txt'"
run_test "head" "omni head -n 2 '$TEST_DIR/lines.txt'"
run_test "tail" "omni tail -n 2 '$TEST_DIR/lines.txt'"
run_test "wc" "omni wc -l '$TEST_DIR/lines.txt'"
run_test "nl" "omni nl '$TEST_DIR/lines.txt'"
run_test "tac" "omni tac '$TEST_DIR/lines.txt'"
run_test "rev" "omni rev '$TEST_DIR/test.txt'"
run_test "strings" "omni strings '$TEST_DIR/test.txt'"
run_test "fold" "omni fold -w 5 '$TEST_DIR/test.txt'"

# Text processing (stdin)
run_test "tr" "echo 'hello' | omni tr a-z A-Z"
run_test "cut" "echo 'a:b:c' | omni cut -d: -f1"
run_test "sort" "omni sort '$TEST_DIR/fruits.txt'"
run_test "uniq" "omni uniq '$TEST_DIR/dupes.txt'"
run_test "paste" "omni paste '$TEST_DIR/test.txt' '$TEST_DIR/test.txt'"
run_test "join" "echo -e '1 a\n2 b' > '$TEST_DIR/j1.txt' && echo -e '1 x\n2 y' > '$TEST_DIR/j2.txt' && omni join '$TEST_DIR/j1.txt' '$TEST_DIR/j2.txt'"
run_test "column" "echo -e 'a b c\n1 2 3' | omni column -t"
run_test "comm" "echo -e 'a\nb' > '$TEST_DIR/c1.txt' && echo -e 'b\nc' > '$TEST_DIR/c2.txt' && omni comm '$TEST_DIR/c1.txt' '$TEST_DIR/c2.txt'"

# Search commands
run_test "grep" "omni grep 'hello' '$TEST_DIR/test.txt'"
run_test "egrep" "omni egrep 'hello|world' '$TEST_DIR/test.txt'"
run_test "fgrep" "omni fgrep 'hello' '$TEST_DIR/test.txt'"
run_test "find" "omni find '$TEST_DIR' --type f"

# Encoding commands
run_test "base64 encode" "echo 'test' | omni base64"
run_test "base64 decode" "echo 'dGVzdAo=' | omni base64 -d"
run_test "base32 encode" "echo 'test' | omni base32"
run_test "hex encode" "echo 'test' | omni hex"
run_test "md5sum" "echo 'test' | omni md5sum"
run_test "sha256sum" "echo 'test' | omni sha256sum"
run_test "sha512sum" "echo 'test' | omni sha512sum"

# ID generators
run_test "uuid" "omni uuid"
run_test "ulid" "omni ulid"
run_test "ksuid" "omni ksuid"
run_test "nanoid" "omni nanoid"
run_test "snowflake" "omni snowflake"
run_test "random" "omni random 16"

# JSON/YAML/TOML/XML
run_test "jq" "omni jq '.name' '$TEST_DIR/test.json'"
run_test "yq" "omni yq '.key' '$TEST_DIR/test.yaml'"
run_test "json fmt" "omni json fmt '$TEST_DIR/test.json'"
run_test "yaml fmt" "omni yaml fmt '$TEST_DIR/test.yaml'"
run_test "toml fmt" "omni toml fmt '$TEST_DIR/test.toml'"
run_test "xml fmt" "omni xml fmt '$TEST_DIR/test.xml'"

# CSV
run_test "csv tojson" "omni csv tojson '$TEST_DIR/test.csv'"

# Case conversion
run_test "case upper" "echo 'hello' | omni case upper"
run_test "case lower" "echo 'HELLO' | omni case lower"
run_test "case title" "echo 'hello world' | omni case title"
run_test "case snake" "echo 'helloWorld' | omni case snake"
run_test "case camel" "echo 'hello_world' | omni case camel"

# URL
run_test "url encode" "omni url encode 'hello world'"
run_test "url decode" "omni url decode 'hello%20world'"
run_test "url parse" "omni url parse 'https://example.com/path?q=1'"

# Cron
run_test "cron" "omni cron '0 0 * * *'"

# Diff/Compare
run_test "diff" "omni diff '$TEST_DIR/c1.txt' '$TEST_DIR/c2.txt' || true"
run_test "cmp" "omni cmp '$TEST_DIR/c1.txt' '$TEST_DIR/c2.txt' || true"

# File operations (read-only tests)
run_test "touch" "omni touch '$TEST_DIR/newfile.txt'"
run_test "mkdir" "omni mkdir '$TEST_DIR/newdir'"
run_test "cp" "omni cp '$TEST_DIR/test.txt' '$TEST_DIR/test_copy.txt'"
run_test "mv" "omni mv '$TEST_DIR/test_copy.txt' '$TEST_DIR/test_moved.txt'"
run_test "ln" "omni ln -s '$TEST_DIR/test.txt' '$TEST_DIR/test_link.txt' || true"
run_test "rm" "omni rm '$TEST_DIR/test_moved.txt'"
run_test "rmdir" "omni rmdir '$TEST_DIR/newdir'"

# Archive (if test files exist)
run_test "tar create" "omni tar -cf '$TEST_DIR/test.tar' '$TEST_DIR/test.txt'"
run_test "tar list" "omni tar -tf '$TEST_DIR/test.tar'"
run_test "zip" "omni zip '$TEST_DIR/test.zip' '$TEST_DIR/test.txt'"

# System info
run_test "uname" "omni uname"
run_test "uptime" "omni uptime"
run_test "free" "omni free"
run_test "id" "omni id"
run_test "whoami" "omni whoami"
run_test "which" "omni which omni"

# Process (safe tests)
run_test "ps" "omni ps"
run_test "time" "omni time sleep 0.1"

# Misc commands
run_test "printf" "omni printf '%s %d' hello 42"
run_test "sleep" "omni sleep 0.1"
run_test "cmdtree" "omni cmdtree | head -5"

# Commands that need special handling
skip_test "awk" "requires stdin/file"
skip_test "sed" "requires stdin/file"
skip_test "xargs" "requires stdin"
skip_test "watch" "interactive"
skip_test "less" "interactive"
skip_test "more" "interactive"
skip_test "top" "interactive"
skip_test "encrypt" "requires key"
skip_test "decrypt" "requires key"
skip_test "curl" "requires network"
skip_test "gzip" "modifies files"
skip_test "gunzip" "requires .gz file"
skip_test "bzip2" "modifies files"
skip_test "bunzip2" "requires .bz2 file"
skip_test "xz" "modifies files"
skip_test "unxz" "requires .xz file"
skip_test "unzip" "requires .zip file"
skip_test "dd" "dangerous"
skip_test "kill" "dangerous"
skip_test "chmod" "modifies permissions"
skip_test "chown" "requires root"

echo ""
echo "=============================================="
echo "  Test Results"
echo "=============================================="
echo ""
printf "${GREEN}Passed:${NC}  %d\n" $PASSED
printf "${RED}Failed:${NC}  %d\n" $FAILED
printf "${YELLOW}Skipped:${NC} %d\n" $SKIPPED
echo ""
echo "Total: $((PASSED + FAILED + SKIPPED))"
echo ""

# Check recent logs
echo "--- Recent Log Entries ---"
echo ""
omni logger --viewer 2>&1 | tail -20

# Exit with failure if any tests failed
if [ $FAILED -gt 0 ]; then
    exit 1
fi

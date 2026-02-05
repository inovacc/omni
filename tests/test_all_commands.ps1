#Requires -Version 5.1
<#
.SYNOPSIS
    Comprehensive test script for omni - all commands and features

.DESCRIPTION
    Tests all omni commands including core, file, text, system, archive, hash,
    encoding, data processing, utilities, and more.

.PARAMETER Category
    Test category to run. Options: core, file, text, system, archive, hash,
    encoding, data, util, id, security, generate, buf, db, pager, special, stdin, error, all

.PARAMETER Verbose
    Show verbose output including command being run

.PARAMETER OmniBinary
    Path to omni binary (default: .\omni.exe)

.EXAMPLE
    .\test_all_commands.ps1 -Category all
    .\test_all_commands.ps1 -Category core -Verbose
#>

param(
    [ValidateSet("core", "file", "text", "system", "archive", "hash", "encoding",
                 "data", "util", "id", "security", "generate", "buf", "db",
                 "pager", "special", "stdin", "error", "all")]
    [string]$Category = "all",

    [switch]$VerboseOutput,

    [string]$OmniBinary = ".\omni.exe"
)

$ErrorActionPreference = "Stop"

# Counters
$script:Passed = 0
$script:Failed = 0
$script:Skipped = 0
$script:Total = 0

# Colors
function Write-Pass { param($msg) Write-Host "[PASS] $msg" -ForegroundColor Green; $script:Passed++; $script:Total++ }
function Write-Fail { param($msg, $err) Write-Host "[FAIL] $msg" -ForegroundColor Red; if ($VerboseOutput -and $err) { Write-Host "       Error: $err" -ForegroundColor Red }; $script:Failed++; $script:Total++ }
function Write-Skip { param($msg) Write-Host "[SKIP] $msg" -ForegroundColor Yellow; $script:Skipped++; $script:Total++ }
function Write-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Blue }
function Write-Section {
    param($msg)
    Write-Host ""
    Write-Host ("=" * 78) -ForegroundColor Cyan
    Write-Host "  $msg" -ForegroundColor Cyan
    Write-Host ("=" * 78) -ForegroundColor Cyan
}

# Test helper - run command and check exit code
function Test-Command {
    param(
        [string]$Name,
        [string]$Command,
        [int]$ExpectedExitCode = 0
    )

    if ($VerboseOutput) {
        Write-Host "  Running: $Command" -ForegroundColor Blue
    }

    try {
        $output = Invoke-Expression $Command 2>&1
        $exitCode = $LASTEXITCODE
        if ($null -eq $exitCode) { $exitCode = 0 }

        if ($exitCode -eq $ExpectedExitCode) {
            Write-Pass $Name
            return $true
        } else {
            Write-Fail $Name "Exit code $exitCode (expected $ExpectedExitCode): $output"
            return $false
        }
    } catch {
        Write-Fail $Name $_.Exception.Message
        return $false
    }
}

# Test helper - run command and check output contains string
function Test-CommandContains {
    param(
        [string]$Name,
        [string]$Command,
        [string]$Expected
    )

    if ($VerboseOutput) {
        Write-Host "  Running: $Command" -ForegroundColor Blue
    }

    try {
        $output = Invoke-Expression $Command 2>&1 | Out-String
        $exitCode = $LASTEXITCODE
        if ($null -eq $exitCode) { $exitCode = 0 }

        if ($exitCode -eq 0 -and $output -match [regex]::Escape($Expected)) {
            Write-Pass $Name
            return $true
        } else {
            Write-Fail $Name "Output does not contain '$Expected': $output"
            return $false
        }
    } catch {
        Write-Fail $Name $_.Exception.Message
        return $false
    }
}

# Test helper - run command and check exact output
function Test-CommandExact {
    param(
        [string]$Name,
        [string]$Command,
        [string]$Expected
    )

    if ($VerboseOutput) {
        Write-Host "  Running: $Command" -ForegroundColor Blue
    }

    try {
        $output = (Invoke-Expression $Command 2>&1 | Out-String).Trim()
        $exitCode = $LASTEXITCODE
        if ($null -eq $exitCode) { $exitCode = 0 }

        if ($exitCode -eq 0 -and $output -eq $Expected) {
            Write-Pass $Name
            return $true
        } else {
            Write-Fail $Name "Expected '$Expected', got '$output'"
            return $false
        }
    } catch {
        Write-Fail $Name $_.Exception.Message
        return $false
    }
}

# Create test directory and fixtures
$TestDir = Join-Path $env:TEMP "omni_test_$(Get-Random)"
New-Item -ItemType Directory -Path $TestDir -Force | Out-Null

function New-Fixtures {
    Write-Info "Creating test fixtures in $TestDir"

    # Text files
    "Hello World`nFoo Bar`nBaz Qux" | Set-Content "$TestDir\test.txt" -NoNewline
    "line1`nline2`nline3`nline4`nline5" | Set-Content "$TestDir\lines.txt" -NoNewline
    "apple`nbanana`napple`ncherry`nbanana`nbanana" | Set-Content "$TestDir\duplicates.txt" -NoNewline
    "3`n1`n4`n1`n5`n9`n2`n6" | Set-Content "$TestDir\numbers.txt" -NoNewline
    "name,age,city`nAlice,30,NYC`nBob,25,LA" | Set-Content "$TestDir\data.csv" -NoNewline
    "col1`tcol2`tcol3`nval1`tval2`tval3" | Set-Content "$TestDir\tabs.txt" -NoNewline
    "  spaces  `n  around  " | Set-Content "$TestDir\spaces.txt" -NoNewline
    "short" | Set-Content "$TestDir\short.txt" -NoNewline
    "This is a very long line that should be folded when using the fold command with a specific width parameter set" | Set-Content "$TestDir\longline.txt" -NoNewline

    # JSON files
    @'
{
    "name": "test",
    "version": "1.0.0",
    "items": [1, 2, 3],
    "nested": {
        "key": "value"
    }
}
'@ | Set-Content "$TestDir\test.json"

    '[{"name":"Alice","age":30},{"name":"Bob","age":25}]' | Set-Content "$TestDir\array.json" -NoNewline

    # YAML files
    @'
name: test
version: 1.0.0
items:
  - 1
  - 2
  - 3
nested:
  key: value
'@ | Set-Content "$TestDir\test.yaml"

    # XML files
    @'
<?xml version="1.0" encoding="UTF-8"?>
<root>
    <name>test</name>
    <items>
        <item>1</item>
        <item>2</item>
    </items>
</root>
'@ | Set-Content "$TestDir\test.xml"

    # TOML files
    @'
name = "test"
version = "1.0.0"

[database]
host = "localhost"
port = 5432
'@ | Set-Content "$TestDir\test.toml"

    # SQL files
    "SELECT id, name, email FROM users WHERE active = true ORDER BY name;" | Set-Content "$TestDir\test.sql" -NoNewline

    # HTML files
    @'
<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<h1>Hello World</h1>
<p>This is a test.</p>
</body>
</html>
'@ | Set-Content "$TestDir\test.html"

    # CSS files
    @'
body {
    background: #fff;
    color: #333;
}
h1 {
    font-size: 24px;
}
'@ | Set-Content "$TestDir\test.css"

    # .env file
    @'
DATABASE_URL=postgres://localhost/test
API_KEY=secret123
DEBUG=true
'@ | Set-Content "$TestDir\test.env"

    # Binary file
    [byte[]]$bytes = 0,1,2,3,4,5
    [System.IO.File]::WriteAllBytes("$TestDir\binary.bin", $bytes)

    # Go source file
    @'
package sample

func Add(a, b int) int {
    return a + b
}

func Subtract(a, b int) int {
    return a - b
}
'@ | Set-Content "$TestDir\sample.go"

    # Create subdirectory structure
    New-Item -ItemType Directory -Path "$TestDir\subdir\nested" -Force | Out-Null
    "nested file" | Set-Content "$TestDir\subdir\nested\file.txt" -NoNewline
    "subdir file" | Set-Content "$TestDir\subdir\file.txt" -NoNewline

    # Proto file
    @'
syntax = "proto3";

package test;

message User {
    string name = 1;
    int32 age = 2;
}
'@ | Set-Content "$TestDir\test.proto"

    Write-Info "Fixtures created"
}

# ============================================================================
# CORE COMMANDS
# ============================================================================
function Test-CoreCommands {
    Write-Section "CORE COMMANDS"

    # echo
    Test-CommandExact "echo: basic" "$OmniBinary echo 'Hello World'" "Hello World"
    Test-CommandContains "echo: -n (no newline)" "$OmniBinary echo -n 'test'" "test"

    # printf
    Test-CommandContains "printf: basic" "$OmniBinary printf '%s %d' hello 42" "hello 42"

    # pwd
    Test-Command "pwd: basic" "$OmniBinary pwd"

    # ls
    Test-Command "ls: current dir" "$OmniBinary ls"
    Test-Command "ls: specific dir" "$OmniBinary ls '$TestDir'"
    Test-CommandContains "ls: -l (long)" "$OmniBinary ls -l '$TestDir'" "test.txt"
    Test-Command "ls: -a (all)" "$OmniBinary ls -a '$TestDir'"
    Test-Command "ls: -R (recursive)" "$OmniBinary ls -R '$TestDir'"

    # cat
    Test-CommandContains "cat: single file" "$OmniBinary cat '$TestDir\test.txt'" "Hello World"
    Test-Command "cat: multiple files" "$OmniBinary cat '$TestDir\test.txt' '$TestDir\lines.txt'"
    Test-CommandContains "cat: -n (number lines)" "$OmniBinary cat -n '$TestDir\test.txt'" "1"

    # tree
    Test-Command "tree: basic" "$OmniBinary tree '$TestDir'"
    Test-Command "tree: -L (level)" "$OmniBinary tree -L 1 '$TestDir'"

    # date
    Test-Command "date: basic" "$OmniBinary date"
    Test-Command "date: +format" "$OmniBinary date '+%Y-%m-%d'"

    # dirname
    Test-CommandContains "dirname: basic" "$OmniBinary dirname /path/to/file.txt" "/path/to"

    # basename
    Test-CommandExact "basename: basic" "$OmniBinary basename /path/to/file.txt" "file.txt"

    # arch
    Test-Command "arch: basic" "$OmniBinary arch"

    # uname
    Test-Command "uname: basic" "$OmniBinary uname"
    Test-Command "uname: -a (all)" "$OmniBinary uname -a"

    # seq
    Test-CommandContains "seq: basic" "$OmniBinary seq 5" "1"
    Test-CommandContains "seq: range" "$OmniBinary seq 2 5" "2"
}

# ============================================================================
# FILE OPERATIONS
# ============================================================================
function Test-FileCommands {
    Write-Section "FILE OPERATIONS"

    # touch
    Test-Command "touch: create file" "$OmniBinary touch '$TestDir\newfile.txt'"
    Test-Command "touch: update existing" "$OmniBinary touch '$TestDir\test.txt'"

    # mkdir
    Test-Command "mkdir: basic" "$OmniBinary mkdir '$TestDir\newdir'"
    Test-Command "mkdir: -p (parents)" "$OmniBinary mkdir -p '$TestDir\deep\nested\dir'"

    # cp
    Test-Command "cp: file" "$OmniBinary cp '$TestDir\test.txt' '$TestDir\test_copy.txt'"
    Test-Command "cp: -r (recursive)" "$OmniBinary cp -r '$TestDir\subdir' '$TestDir\subdir_copy'"

    # mv
    Test-Command "mv: rename" "$OmniBinary mv '$TestDir\newfile.txt' '$TestDir\renamed.txt'"

    # rm
    Test-Command "rm: file" "$OmniBinary rm '$TestDir\renamed.txt'"
    Test-Command "rm: -r (recursive)" "$OmniBinary rm -r '$TestDir\subdir_copy'"

    # rmdir
    Test-Command "rmdir: empty dir" "$OmniBinary rmdir '$TestDir\newdir'"

    # stat
    Test-Command "stat: basic" "$OmniBinary stat '$TestDir\test.txt'"

    # file
    Test-CommandContains "file: text" "$OmniBinary file '$TestDir\test.txt'" "text"

    # find
    Test-Command "find: basic" "$OmniBinary find '$TestDir'"
    Test-CommandContains "find: -name" "$OmniBinary find '$TestDir' -name '*.txt'" "test.txt"

    # which
    Test-Command "which: basic" "$OmniBinary which cmd" 0
}

# ============================================================================
# TEXT PROCESSING
# ============================================================================
function Test-TextCommands {
    Write-Section "TEXT PROCESSING"

    # head
    Test-Command "head: basic" "$OmniBinary head '$TestDir\lines.txt'"
    Test-CommandContains "head: -n 2" "$OmniBinary head -n 2 '$TestDir\lines.txt'" "line1"

    # tail
    Test-Command "tail: basic" "$OmniBinary tail '$TestDir\lines.txt'"
    Test-CommandContains "tail: -n 2" "$OmniBinary tail -n 2 '$TestDir\lines.txt'" "line5"

    # wc
    Test-Command "wc: basic" "$OmniBinary wc '$TestDir\test.txt'"
    Test-Command "wc: -l (lines)" "$OmniBinary wc -l '$TestDir\lines.txt'"
    Test-Command "wc: -w (words)" "$OmniBinary wc -w '$TestDir\test.txt'"

    # sort
    Test-Command "sort: basic" "$OmniBinary sort '$TestDir\duplicates.txt'"
    Test-Command "sort: -r (reverse)" "$OmniBinary sort -r '$TestDir\duplicates.txt'"
    Test-Command "sort: -n (numeric)" "$OmniBinary sort -n '$TestDir\numbers.txt'"

    # uniq
    Test-Command "uniq: basic" "'a`na`nb`nb`nc' | $OmniBinary uniq"

    # grep
    Test-CommandContains "grep: basic" "$OmniBinary grep 'Hello' '$TestDir\test.txt'" "Hello"
    Test-Command "grep: -i (ignore case)" "$OmniBinary grep -i 'hello' '$TestDir\test.txt'"
    Test-Command "grep: -v (invert)" "$OmniBinary grep -v 'Hello' '$TestDir\test.txt'"
    Test-Command "grep: -n (line numbers)" "$OmniBinary grep -n 'line' '$TestDir\lines.txt'"

    # cut
    Test-Command "cut: -d -f (delimiter, field)" "$OmniBinary cut -d',' -f1 '$TestDir\data.csv'"

    # tr
    Test-Command "tr: basic" "'hello' | $OmniBinary tr 'a-z' 'A-Z'"

    # sed
    Test-Command "sed: substitute" "'hello world' | $OmniBinary sed 's/world/universe/'"

    # awk
    Test-Command "awk: print field" "'a b c' | $OmniBinary awk '{print `$2}'"

    # nl
    Test-Command "nl: basic" "$OmniBinary nl '$TestDir\lines.txt'"

    # tac
    Test-CommandContains "tac: reverse lines" "$OmniBinary tac '$TestDir\lines.txt'" "line5"

    # rev
    Test-Command "rev: reverse chars" "'hello' | $OmniBinary rev"

    # column
    Test-Command "column: basic" "$OmniBinary column '$TestDir\data.csv'"

    # fold
    Test-Command "fold: basic" "$OmniBinary fold '$TestDir\longline.txt'"

    # diff
    Test-Command "diff: identical" "$OmniBinary diff '$TestDir\test.txt' '$TestDir\test.txt'"

    # shuf
    Test-Command "shuf: basic" "$OmniBinary shuf '$TestDir\lines.txt'"
}

# ============================================================================
# SYSTEM COMMANDS
# ============================================================================
function Test-SystemCommands {
    Write-Section "SYSTEM COMMANDS"

    # env
    Test-Command "env: basic" "$OmniBinary env"

    # whoami
    Test-Command "whoami: basic" "$OmniBinary whoami"

    # id
    Test-Command "id: basic" "$OmniBinary id"

    # uptime
    Test-Command "uptime: basic" "$OmniBinary uptime"

    # df
    Test-Command "df: basic" "$OmniBinary df"
    Test-Command "df: -h (human)" "$OmniBinary df -h"

    # du
    Test-Command "du: basic" "$OmniBinary du '$TestDir'"
    Test-Command "du: -h (human)" "$OmniBinary du -h '$TestDir'"

    # ps
    Test-Command "ps: basic" "$OmniBinary ps"

    # kill (just test --list)
    Test-Command "kill: -l (list signals)" "$OmniBinary kill -l"
}

# ============================================================================
# ARCHIVE & COMPRESSION
# ============================================================================
function Test-ArchiveCommands {
    Write-Section "ARCHIVE & COMPRESSION"

    "archive test content" | Set-Content "$TestDir\archive_test.txt" -NoNewline

    # tar
    Test-Command "tar: create" "$OmniBinary tar -cf '$TestDir\test.tar' -C '$TestDir' archive_test.txt"
    Test-Command "tar: list" "$OmniBinary tar -tf '$TestDir\test.tar'"
    Test-Command "tar: create gzip" "$OmniBinary tar -czf '$TestDir\test.tar.gz' -C '$TestDir' archive_test.txt"

    # zip
    Test-Command "zip: create" "$OmniBinary zip '$TestDir\test.zip' '$TestDir\archive_test.txt'"

    # unzip
    Test-Command "unzip: list" "$OmniBinary unzip -l '$TestDir\test.zip'"

    # gzip
    Copy-Item "$TestDir\archive_test.txt" "$TestDir\gzip_test.txt"
    Test-Command "gzip: compress" "$OmniBinary gzip '$TestDir\gzip_test.txt'"

    # gunzip
    Test-Command "gunzip: decompress" "$OmniBinary gunzip '$TestDir\gzip_test.txt.gz'"
}

# ============================================================================
# HASH & ENCODING
# ============================================================================
function Test-HashCommands {
    Write-Section "HASH & ENCODING"

    # hash
    Test-Command "hash: sha256" "'test' | $OmniBinary hash sha256"
    Test-Command "hash: sha512" "'test' | $OmniBinary hash sha512"
    Test-Command "hash: md5" "'test' | $OmniBinary hash md5"
    Test-Command "hash: file" "$OmniBinary hash sha256 '$TestDir\test.txt'"

    # sha256sum
    Test-Command "sha256sum: basic" "$OmniBinary sha256sum '$TestDir\test.txt'"

    # md5sum
    Test-Command "md5sum: basic" "$OmniBinary md5sum '$TestDir\test.txt'"
}

function Test-EncodingCommands {
    Write-Section "ENCODING"

    # base64
    Test-Command "base64: encode" "'hello' | $OmniBinary base64"
    Test-Command "base64: decode" "'aGVsbG8=' | $OmniBinary base64 -d"

    # base32
    Test-Command "base32: encode" "'hello' | $OmniBinary base32"

    # base58
    Test-Command "base58: encode" "'hello' | $OmniBinary base58"

    # hex
    Test-Command "hex: encode" "'hello' | $OmniBinary hex encode"
    Test-Command "hex: decode" "'68656c6c6f' | $OmniBinary hex decode"

    # url
    Test-Command "url: encode" "$OmniBinary url encode 'hello world'"
    Test-Command "url: decode" "$OmniBinary url decode 'hello%20world'"

    # html encode/decode
    Test-Command "html: encode" "$OmniBinary html encode '<div>test</div>'"
    Test-Command "html: decode" "$OmniBinary html decode '&lt;div&gt;'"

    # xxd
    Test-CommandContains "xxd: basic dump" "'hello' | $OmniBinary xxd" "68656c6c6f"
    Test-CommandContains "xxd: with address" "'hello' | $OmniBinary xxd" "00000000:"
    Test-Command "xxd: plain mode" "'hello' | $OmniBinary xxd -p"
    Test-Command "xxd: include mode" "'Hi' | $OmniBinary xxd -i"
    Test-Command "xxd: binary mode" "'A' | $OmniBinary xxd -b"
    Test-Command "xxd: uppercase" "'hello' | $OmniBinary xxd -u"
    Test-Command "xxd: custom columns" "'hello world' | $OmniBinary xxd -c 8"
    Test-Command "xxd: reverse plain" "'68656c6c6f' | $OmniBinary xxd -r -p"
    Test-Command "xxd: file" "$OmniBinary xxd '$TestDir\test.txt'"
}

# ============================================================================
# DATA PROCESSING
# ============================================================================
function Test-DataCommands {
    Write-Section "DATA PROCESSING"

    # jq
    Test-CommandContains "jq: query" "$OmniBinary jq '.name' '$TestDir\test.json'" "test"
    Test-Command "jq: array" "$OmniBinary jq '.items[0]' '$TestDir\test.json'"
    Test-Command "jq: nested" "$OmniBinary jq '.nested.key' '$TestDir\test.json'"

    # yq
    Test-CommandContains "yq: query" "$OmniBinary yq '.name' '$TestDir\test.yaml'" "test"

    # json fmt
    Test-Command "json: fmt" "$OmniBinary json fmt '$TestDir\test.json'"
    Test-Command "json: minify" "$OmniBinary json minify '$TestDir\test.json'"
    Test-Command "json: validate" "$OmniBinary json validate '$TestDir\test.json'"
    Test-Command "json: stats" "$OmniBinary json stats '$TestDir\test.json'"
    Test-Command "json: keys" "$OmniBinary json keys '$TestDir\test.json'"
    Test-Command "json: tostruct" "$OmniBinary json tostruct '$TestDir\test.json'"
    Test-Command "json: tocsv" "$OmniBinary json tocsv '$TestDir\array.json'"
    Test-Command "json: fromcsv" "$OmniBinary json fromcsv '$TestDir\data.csv'"
    Test-Command "json: toyaml" "$OmniBinary json toyaml '$TestDir\test.json'"

    # yaml
    Test-Command "yaml: validate" "$OmniBinary yaml validate '$TestDir\test.yaml'"
    Test-Command "yaml: fmt" "$OmniBinary yaml fmt '$TestDir\test.yaml'"
    Test-Command "yaml: tostruct" "$OmniBinary yaml tostruct '$TestDir\test.yaml'"

    # xml
    Test-Command "xml: fmt" "$OmniBinary xml fmt '$TestDir\test.xml'"
    Test-Command "xml: validate" "$OmniBinary xml validate '$TestDir\test.xml'"
    Test-Command "xml: tojson" "$OmniBinary xml tojson '$TestDir\test.xml'"

    # toml
    Test-Command "toml: validate" "$OmniBinary toml validate '$TestDir\test.toml'"
    Test-Command "toml: fmt" "$OmniBinary toml fmt '$TestDir\test.toml'"

    # sql
    Test-Command "sql: fmt" "$OmniBinary sql fmt '$TestDir\test.sql'"
    Test-Command "sql: minify" "$OmniBinary sql minify '$TestDir\test.sql'"
    Test-Command "sql: validate" "$OmniBinary sql validate '$TestDir\test.sql'"

    # html
    Test-Command "html: fmt" "$OmniBinary html fmt '$TestDir\test.html'"
    Test-Command "html: minify" "$OmniBinary html minify '$TestDir\test.html'"
    Test-Command "html: validate" "$OmniBinary html validate '$TestDir\test.html'"

    # css
    Test-Command "css: fmt" "$OmniBinary css fmt '$TestDir\test.css'"
    Test-Command "css: minify" "$OmniBinary css minify '$TestDir\test.css'"
    Test-Command "css: validate" "$OmniBinary css validate '$TestDir\test.css'"

    # dotenv
    Test-Command "dotenv: basic" "$OmniBinary dotenv '$TestDir\test.env'"
    Test-CommandContains "dotenv: get" "$OmniBinary dotenv '$TestDir\test.env' API_KEY" "secret123"
}

# ============================================================================
# ID GENERATORS
# ============================================================================
function Test-IdGenerators {
    Write-Section "ID GENERATORS"

    # uuid
    Test-Command "uuid: v4" "$OmniBinary uuid"
    Test-Command "uuid: v1" "$OmniBinary uuid v1"
    Test-Command "uuid: v7" "$OmniBinary uuid v7"

    # random
    Test-Command "random: string" "$OmniBinary random string 16"
    Test-Command "random: hex" "$OmniBinary random hex 16"
    Test-Command "random: int" "$OmniBinary random int 1 100"
    Test-Command "random: password" "$OmniBinary random password 16"

    # ksuid
    Test-Command "ksuid: generate" "$OmniBinary ksuid"

    # ulid
    Test-Command "ulid: generate" "$OmniBinary ulid"

    # snowflake
    Test-Command "snowflake: generate" "$OmniBinary snowflake"

    # nanoid
    Test-Command "nanoid: generate" "$OmniBinary nanoid"
}

# ============================================================================
# SECURITY
# ============================================================================
function Test-SecurityCommands {
    Write-Section "SECURITY"

    # encrypt/decrypt
    Test-Command "encrypt: basic" "'secret data' | $OmniBinary encrypt -p 'password123' > '$TestDir\encrypted.bin'"
    Test-Command "decrypt: basic" "$OmniBinary decrypt -p 'password123' '$TestDir\encrypted.bin'"

    # jwt
    $jwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
    Test-Command "jwt: decode" "$OmniBinary jwt decode '$jwt'"
}

# ============================================================================
# UTILITIES
# ============================================================================
function Test-UtilityCommands {
    Write-Section "UTILITIES"

    # case
    Test-Command "case: upper" "$OmniBinary case upper 'hello world'"
    Test-Command "case: lower" "$OmniBinary case lower 'HELLO WORLD'"
    Test-Command "case: title" "$OmniBinary case title 'hello world'"
    Test-Command "case: snake" "$OmniBinary case snake 'helloWorld'"
    Test-Command "case: camel" "$OmniBinary case camel 'hello_world'"

    # cron
    Test-Command "cron: parse" "$OmniBinary cron parse '0 * * * *'"
    Test-Command "cron: next" "$OmniBinary cron next '0 * * * *'"

    # loc
    Test-Command "loc: basic" "$OmniBinary loc '$TestDir'"

    # cmdtree
    Test-Command "cmdtree: basic" "$OmniBinary cmdtree"
}

# ============================================================================
# CODE GENERATION
# ============================================================================
function Test-GenerateCommands {
    Write-Section "CODE GENERATION"

    try { Test-Command "generate: test" "$OmniBinary generate test '$TestDir\sample.go'" } catch { Write-Skip "generate test: not available" }
    try { Test-Command "generate: handler" "$OmniBinary generate handler User" } catch { Write-Skip "generate handler: not available" }
}

# ============================================================================
# ERROR HANDLING
# ============================================================================
function Test-ErrorHandling {
    Write-Section "ERROR HANDLING"

    Test-Command "error: nonexistent file" "$OmniBinary cat '$TestDir\nonexistent.txt'" 1
    Test-Command "error: invalid json" "'not json' | $OmniBinary json validate" 1
}

# ============================================================================
# MAIN
# ============================================================================

Write-Host ""
Write-Host ("=" * 78) -ForegroundColor Cyan
Write-Host "              OMNI - Comprehensive Command Test Suite (PowerShell)" -ForegroundColor Cyan
Write-Host ("=" * 78) -ForegroundColor Cyan
Write-Host ""

# Check if omni exists
if (-not (Test-Path $OmniBinary)) {
    Write-Info "Building omni..."
    & go build -o $OmniBinary .
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build omni"
        exit 1
    }
}

Write-Info "Using omni binary: $OmniBinary"
Write-Info "Test directory: $TestDir"
Write-Info "Category: $Category"
Write-Info "Verbose: $VerboseOutput"
Write-Host ""

New-Fixtures

switch ($Category) {
    "core" { Test-CoreCommands }
    "file" { Test-FileCommands }
    "text" { Test-TextCommands }
    "system" { Test-SystemCommands }
    "archive" { Test-ArchiveCommands }
    "hash" { Test-HashCommands }
    "encoding" { Test-EncodingCommands }
    "data" { Test-DataCommands }
    "util" { Test-UtilityCommands }
    "id" { Test-IdGenerators }
    "security" { Test-SecurityCommands }
    "generate" { Test-GenerateCommands }
    "error" { Test-ErrorHandling }
    "all" {
        Test-CoreCommands
        Test-FileCommands
        Test-TextCommands
        Test-SystemCommands
        Test-ArchiveCommands
        Test-HashCommands
        Test-EncodingCommands
        Test-DataCommands
        Test-IdGenerators
        Test-SecurityCommands
        Test-UtilityCommands
        Test-GenerateCommands
        Test-ErrorHandling
    }
    default {
        Write-Error "Unknown category: $Category"
        exit 1
    }
}

# Cleanup
Remove-Item -Path $TestDir -Recurse -Force -ErrorAction SilentlyContinue

# Summary
Write-Host ""
Write-Host ("=" * 78) -ForegroundColor Cyan
Write-Host "  TEST SUMMARY" -ForegroundColor Cyan
Write-Host ("=" * 78) -ForegroundColor Cyan
Write-Host ""
Write-Host "  Total:   $script:Total"
Write-Host "  Passed:  $script:Passed" -ForegroundColor Green
Write-Host "  Failed:  $script:Failed" -ForegroundColor Red
Write-Host "  Skipped: $script:Skipped" -ForegroundColor Yellow
Write-Host ""

if ($script:Failed -gt 0) {
    Write-Host "Some tests failed!" -ForegroundColor Red
    exit 1
} else {
    Write-Host "All tests passed!" -ForegroundColor Green
    exit 0
}

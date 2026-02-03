# Compare ripgrep (rg) vs omni rg behavior
# PowerShell version for Windows

$ErrorActionPreference = "Stop"

# Counters
$script:PASS = 0
$script:FAIL = 0
$script:SKIP = 0

# Paths
$ScriptPath = $MyInvocation.MyCommand.Path
if ($ScriptPath) {
    $ProjectRoot = Split-Path -Parent (Split-Path -Parent $ScriptPath)
} else {
    $ProjectRoot = (Get-Location).Path
}
$Omni = Join-Path $ProjectRoot "omni.exe"

# Build omni if needed
if (-not (Test-Path $Omni)) {
    Write-Host "Building omni..."
    Push-Location $ProjectRoot
    go build -o $Omni .
    Pop-Location
}

# Create temp directory
$TestDir = Join-Path $env:TEMP "rg_compare_$(Get-Random)"
New-Item -ItemType Directory -Path $TestDir -Force | Out-Null

Write-Host "=== ripgrep vs omni rg Comparison Test ===" -ForegroundColor Blue
Write-Host "Test directory: $TestDir"
Write-Host ""

function Create-Fixtures {
    # Go files
    New-Item -ItemType Directory -Path "$TestDir\src\pkg" -Force | Out-Null

    @'
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
'@ | Out-File -FilePath "$TestDir\src\main.go" -Encoding UTF8

    @'
package pkg

// Helper is a helper function
func Helper() string {
    return "helper"
}

// TODO: Add more helpers
// FIXME: This needs improvement
'@ | Out-File -FilePath "$TestDir\src\pkg\helper.go" -Encoding UTF8

    @'
package pkg

import "testing"

func TestHelper(t *testing.T) {
    result := Helper()
    if result != "helper" {
        t.Errorf("expected helper, got %s", result)
    }
}
'@ | Out-File -FilePath "$TestDir\src\pkg\helper_test.go" -Encoding UTF8

    # JavaScript
    New-Item -ItemType Directory -Path "$TestDir\src\js" -Force | Out-Null
    @'
const express = require('express');

function main() {
    console.log("Hello from JavaScript");
}

// TODO: Add error handling
main();
'@ | Out-File -FilePath "$TestDir\src\js\app.js" -Encoding UTF8

    # Python
    New-Item -ItemType Directory -Path "$TestDir\src\py" -Force | Out-Null
    @'
#!/usr/bin/env python3

def main():
    print("Hello from Python")

# TODO: Add logging
if __name__ == "__main__":
    main()
'@ | Out-File -FilePath "$TestDir\src\py\app.py" -Encoding UTF8

    # Markdown
    @'
# Test Project

This is a test project for comparing rg implementations.

## Features

- Feature 1: Hello World
- Feature 2: Error handling

## TODO

- Add more tests
'@ | Out-File -FilePath "$TestDir\README.md" -Encoding UTF8

    # JSON
    @'
{
    "name": "test-project",
    "version": "1.0.0",
    "hello": "world"
}
'@ | Out-File -FilePath "$TestDir\config.json" -Encoding UTF8

    # Patterns file
    @'
hello world
Hello World
HELLO WORLD
func()
func(arg)
test@example.com
192.168.1.1
'@ | Out-File -FilePath "$TestDir\patterns.txt" -Encoding UTF8

    # Gitignore
    @'
*.log
node_modules/
'@ | Out-File -FilePath "$TestDir\.gitignore" -Encoding UTF8

    # Ignored files
    "debug info" | Out-File -FilePath "$TestDir\debug.log" -Encoding UTF8
}

function Compare-Count {
    param(
        [string]$Description,
        [string]$RgCmd,
        [string]$OmniCmd
    )

    try {
        $rgOutput = Invoke-Expression $RgCmd 2>&1 | Out-String
        $rgCount = ($rgOutput -split "`n" | Where-Object { $_.Trim() -ne "" }).Count
    } catch {
        $rgCount = 0
    }

    try {
        $omniOutput = Invoke-Expression $OmniCmd 2>&1 | Out-String
        $omniCount = ($omniOutput -split "`n" | Where-Object { $_.Trim() -ne "" }).Count
    } catch {
        $omniCount = 0
    }

    if ($rgCount -eq $omniCount) {
        Write-Host "[PASS] $Description (both: $rgCount lines)" -ForegroundColor Green
        $script:PASS++
    } else {
        Write-Host "[FAIL] $Description" -ForegroundColor Red
        Write-Host "  rg: $rgCount lines, omni: $omniCount lines"
        $script:FAIL++
    }
}

function Test-Feature {
    param(
        [string]$Description,
        [string]$OmniCmd,
        [string]$Expected
    )

    try {
        $output = Invoke-Expression $OmniCmd 2>&1 | Out-String
    } catch {
        $output = $_.Exception.Message
    }

    if ($output -match [regex]::Escape($Expected)) {
        Write-Host "[PASS] $Description" -ForegroundColor Green
        $script:PASS++
    } else {
        Write-Host "[FAIL] $Description" -ForegroundColor Red
        Write-Host "  Expected: $Expected"
        Write-Host "  Got: $($output.Substring(0, [Math]::Min(100, $output.Length)))"
        $script:FAIL++
    }
}

function Run-Tests {
    Push-Location $TestDir

    Write-Host "`n--- Basic Pattern Matching ---" -ForegroundColor Blue

    Compare-Count "Simple pattern search" `
        "rg 'Hello' ." `
        "& '$Omni' rg 'Hello' ."

    Compare-Count "Case insensitive (-i)" `
        "rg -i 'hello' ." `
        "& '$Omni' rg -i 'hello' ."

    Compare-Count "Word boundary (-w)" `
        "rg -w 'main' ." `
        "& '$Omni' rg -w 'main' ."

    Compare-Count "Fixed string (-F)" `
        "rg -F 'func()' ." `
        "& '$Omni' rg -F 'func()' ."

    Write-Host "`n--- File Type Filtering ---" -ForegroundColor Blue

    Compare-Count "Go files only (-t go)" `
        "rg -t go 'func' ." `
        "& '$Omni' rg -t go 'func' ."

    Compare-Count "JavaScript files (-t js)" `
        "rg -t js 'function' ." `
        "& '$Omni' rg -t js 'function' ."

    Compare-Count "Python files (-t py)" `
        "rg -t py 'def' ." `
        "& '$Omni' rg -t py 'def' ."

    Write-Host "`n--- Output Modes ---" -ForegroundColor Blue

    Compare-Count "Count mode (-c)" `
        "rg -c 'TODO' ." `
        "& '$Omni' rg -c 'TODO' ."

    Compare-Count "Files with matches (-l)" `
        "rg -l 'Hello' ." `
        "& '$Omni' rg -l 'Hello' ."

    Write-Host "`n--- Context Lines ---" -ForegroundColor Blue

    Compare-Count "After context (-A 2)" `
        "rg -A 2 'import' src\main.go" `
        "& '$Omni' rg -A 2 'import' src\main.go"

    Compare-Count "Before context (-B 2)" `
        "rg -B 2 'run' src\main.go" `
        "& '$Omni' rg -B 2 'run' src\main.go"

    Write-Host "`n--- Glob Patterns ---" -ForegroundColor Blue

    Compare-Count "Include glob (-g '*.go')" `
        "rg -g '*.go' 'func' ." `
        "& '$Omni' rg -g '*.go' 'func' ."

    Write-Host "`n--- Omni-specific Features ---" -ForegroundColor Blue

    Test-Feature "JSON output (--json)" `
        "& '$Omni' rg --json 'Hello' ." `
        "total_matches"

    Write-Host "`n--- Edge Cases ---" -ForegroundColor Blue

    Compare-Count "No matches" `
        "rg 'NONEXISTENT_12345' ." `
        "& '$Omni' rg 'NONEXISTENT_12345' ."

    Pop-Location
}

# Run
Create-Fixtures
Run-Tests

# Cleanup
Remove-Item -Recurse -Force $TestDir -ErrorAction SilentlyContinue

# Summary
Write-Host ""
Write-Host "=== Summary ===" -ForegroundColor Blue
Write-Host "Passed: $script:PASS" -ForegroundColor Green
Write-Host "Failed: $script:FAIL" -ForegroundColor Red
Write-Host "Skipped: $script:SKIP" -ForegroundColor Yellow
Write-Host ""

if ($script:FAIL -gt 0) {
    Write-Host "Some tests failed!" -ForegroundColor Red
    exit 1
} else {
    Write-Host "All tests passed!" -ForegroundColor Green
    exit 0
}
